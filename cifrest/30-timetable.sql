-- ==============================================================================================================
-- 30-timetable.sql function to return a list of schedules passing through a specific crs code on a date and hour
--                  of the day.
-- ==============================================================================================================

CREATE OR REPLACE FUNCTION timetable.getservicesbycrs(pcrs CHAR(3),
                                                      pst TIMESTAMP WITH TIME ZONE)
    RETURNS JSON AS
$$
WITH schedules AS (
    -- The schedules that exist for the station on the require hour
    SELECT * FROM timetable.schedules(pcrs, date_trunc('hour', pst), null)
),
     services AS (
         -- view of services reduced down to the ones active on a specific day
         -- distinct on uid as only 1 entry per schedule
         -- we'll presume that circular routes don't visit a station more than once per hour
         SELECT DISTINCT ON (s.uid) s.id    AS sid,
                                    s.uid,
                                    s.startDate,
                                    s.stp,
                                    st.time AS "time"
         FROM timetable.schedule s
                  INNER JOIN schedules st ON st.sid = s.id
              -- order by uid & stp so that the distinct selects the correct one
              -- e.g. first stp in C, N, O P is selected handling overlays/cancellations correctly
         ORDER BY s.uid, s.stp
     ),
     servicesout AS (
         -- the final schedule list with all required data
         SELECT --s.uid,
                --s.startDate,
                --s.stp,
                st.time     AS "time",
                st.ord      AS ord,
                json_build_object(
                        'tiploc', ot.tiploc,
                        'time', ot.tod
                    )       AS origin,
                --ot.tiploc                   AS origin,
                --ot.tod                      AS "originTime",
                json_build_object(
                        'tiploc', dt.tiploc,
                        'time', dt.tod
                    )       AS destination,
                --dt.tiploc                   AS destination,
                --dt.tod                      AS "destinationTime",
                sj.schedule AS "schedule"
         FROM timetable.schedule s
                  INNER JOIN schedules st ON st.sid = s.id
                  INNER JOIN timetable.origin(s.id) ot ON s.id = ot.sid
                  INNER JOIN timetable.destination(s.id) dt ON s.id = dt.sid
                  INNER JOIN timetable.schedule_json sj ON s.id = sj.id
         WHERE s.id IN (SELECT s1.sid from services s1)
           -- Don't allow cancellations in the output as no need to know
           AND s.stp != 'C'
           -- order by time allowing for midnight being in the range
         ORDER BY st.time >= '01:00', st.time, s.id
     ),
     tpls AS (
         -- View of all origin/destination tiplocs in the output
         SELECT t.*
         FROM timetable.tiploc t
                  INNER JOIN servicesout s
                             ON s.origin ->> 'tiploc' = t.tiploc
                                 OR s.destination ->> 'tiploc' = t.tiploc
                                 OR pcrs = t.crs
         UNION
         -- This stations Tiploc
         SELECT t.*
         FROM timetable.tiploc t
         WHERE t.crs = pcrs
     ),
     date1 AS (
         SELECT
             -- The requested hour
             date_trunc('hour', pst)::TIMESTAMP WITHOUT TIME ZONE                        AS ts,
             -- The next hour
             (date_trunc('hour', pst) + '1 hour'::INTERVAL)::TIMESTAMP WITHOUT TIME ZONE AS next,
             -- The previous hour
             (date_trunc('hour', pst) - '1 hour'::INTERVAL)::TIMESTAMP WITHOUT TIME ZONE AS previous,
             -- The timetable start date, earliest to allow is today
             CASE
                 WHEN c.userstartdate > date_trunc('day', NOW()) THEN c.userstartdate
                 ELSE date_trunc('day', NOW())
                 END                                                                     as userstartdate,
             -- The end date
             c.userenddate
         FROM timetable.cif c
         ORDER BY c.id DESC
         LIMIT 1
     ),
     dates AS (
         -- This takes date1 and sets next or prev to null if they are outside the
         -- available data range
         SELECT pcrs                               as crs,
                (SELECT t.tiploc
                 FROM timetable.tiploc t
                 WHERE t.crs = pcrs
                 ORDER BY nlcdesc DESC, stanox, nlc
                 LIMIT 1)                          AS tiploc,
                d.ts,
                CASE
                    WHEN d.next <= d.userenddate THEN d.next
                    ELSE NULL
                    END                            AS next,
                CASE
                    WHEN d.previous > d.userstartdate THEN d.previous
                    ELSE NULL
                    END                            AS previous,
                d.userstartdate::DATE,
                d.userenddate::DATE,
                -- timestamp of when this timetable was generated
                NOW() AT TIME ZONE 'Europe/London' AS generated
         FROM date1 d
     )
SELECT json_build_object(
               'query', (SELECT row_to_json(d) FROM dates d LIMIT 1),
               'schedules', (SELECT json_agg(row_to_json(s)) FROM servicesout s),
               'tiploc',
               (SELECT json_object_agg(t.tiploc, row_to_json(t)) FROM (select distinct * from tpls order by tiploc) t)
           )::JSON;

$$ LANGUAGE SQL;
