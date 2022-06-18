@Library('ystv-jenkins')

String registryEndpoint = 'registry.comp.ystv.co.uk'

def image
String imageName = 'showtime:$env.BRANCH_NAME-$env.BUILD_ID'

pipeline {
  agent {
    label 'docker'
  }

  environment {
    DOCKER_BUILDKIT = '1'
  }

  stages {
    stage('Prepare') {
      steps {
        ciSkip(action: 'check')
      }
    }

    stage('Build image') {
      steps {
        script {
          image = docker.build(imageName)
        }
      }
    }

    stage('Push image to registry') {
      steps {
        script {
          image.push()
          if ( env.BRANCH_IS_PRIMARY ) {
            image.push('latest')
          }
        }
      }
    }

    stage('Deploy') {
      when {
        expression { env.BRANCH_IS_PRIMARY }
      }
      steps {
        build(job: 'Deploy Nomad Job', parameters: [
          string(name: 'JOB_FILE', value: 'showtime.nomad'),
          text(name: 'TAG_REPLACEMENTS', value: '$registryEndpoint/$imageName')
        ])
      }
    }
  }

  post {
    always {
      ciSkip(action: 'postProcess')
    }
  }
}
