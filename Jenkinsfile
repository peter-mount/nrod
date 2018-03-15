// Repository name use, must end with / or be '' for none
repository= 'area51/'

// image prefix
imagePrefix = 'nre-cif'

// The image version, master branch is latest in docker
version=BRANCH_NAME
if( version == 'master' ) {
  version = 'latest'
}

// The architectures to build, in format recognised by docker
architectures = [ 'amd64', 'arm64v8' ]

// Temp docker image name
tempImage = 'temp/' + imagePrefix + ':' + version

// The docker image name
// architecture can be '' for multiarch images
def dockerImage = {
  architecture -> repository + imagePrefix +
    ':' +
    ( architecture=='' ? '' : (architecture + '-') ) +
    version
}

// The multi arch image name
multiImage = repository + imagePrefix + ':' + version

// The go arch
def goarch = {
  architecture -> switch( architecture ) {
    case 'amd64':
      return 'amd64'
    case 'arm32v6':
    case 'arm32v7':
      return 'arm'
    case 'arm64v8':
      return 'arm64'
    default:
      return architecture
  }
}

// goarm is for arm32 only
def goarm = {
  architecture -> switch( architecture ) {
    case 'arm32v6':
      return '6'
    case 'arm32v7':
      return '7'
    default:
      return ''
  }
}

// Build properties
properties([
  buildDiscarder(logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '10')),
  disableConcurrentBuilds(),
  disableResume()
])

// Build a service for a specific architecture
def buildArch = {
  architecture ->
    sh 'docker build' +
      ' -t ' + dockerImage( architecture ) +
      ' --build-arg skipTest=true' +
      ' --build-arg arch=' + architecture +
      ' --build-arg goos=linux' +
      ' --build-arg goarch=' + goarch( architecture ) +
      ' --build-arg goarm=' + goarm( architecture ) +
      ' .'

    if( repository != '' ) {
      // Push all built images relevant docker repository
      sh 'docker push ' + dockerImage( architecture )
    } // repository != ''
}

manifests = architectures.collect { architecture -> dockerImage( architecture ) }
manifests.join(' ')

// Deploy multi-arch image for a service
def multiArchService = {
  tmp ->
    // Create/amend the manifest with our architectures
    sh 'docker manifest create -a ' + multiImage + ' ' + manifests

    // For each architecture annotate them to be correct
    architectures.each {
      architecture -> sh 'docker manifest annotate' +
        ' --os linux' +
        ' --arch ' + goarch( architecture ) +
        ' ' + multiImage +
        ' ' + dockerImage( architecture )
    }

    // Publish the manifest
    sh 'docker manifest push -p ' + multiImage
}

// Now build everything on one node
node('AMD64') {
  stage( "Checkout" ) {
    checkout scm

    // Prepare the go base image with the source and libraries
    sh 'docker pull golang:alpine'

    // Run up to the source target so libraries are checked out
    sh 'docker build -t ' + tempImage + ' --target source .'
  }

  // Run unit tests
  /*
  stage("Run Tests") {
    def runTest = {
      test -> sh 'docker run -i --rm ' + tempImage + ' go test -v ' + test
    }
    parallel (
      'darwind3': { runTest( 'darwind3' ) },
      'darwinref': { runTest( 'darwinref' ) },
      'ldb': { runTest( 'ldb' ) },
      'util': { runTest( 'util' ) }
    )
    //sh 'docker build -t ' + tempImage + ' --target test .'
  }
  */

  stage( 'Build' ) {
    parallel (
      'amd64': {
        buildArch( "amd64" )
      },
      'arm64v8': {
        buildArch( "arm64v8" )
      }
    )
  }

  // Stages valid only if we have a repository set
  if( repository != '' ) {
    stage( "Multiarch Image" ) {
      multiArchService( '' )
    }
  }

}