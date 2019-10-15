pipeline {
  agent any
  options {
    disableConcurrentBuilds()
  }

  stages {
    stage('build') {
      steps {
        checkout scm
        script {
          def nodeInstallRoot = installTool('node-v10.15.1')
          def yarnInstallRoot = installTool('yarn-1_17_3')
          def goInstallRoot = installTool('go1.12.9.linux-amd64')
          configFileProvider([configFile(fileId: 'artifactory-npmrc', targetLocation: '.npmrc')]) {
            sshAsVxPipeline {
              sh(returnStdout: true, script: """
PATH=\$PATH:${nodeInstallRoot}:${nodeInstallRoot}/bin:${yarnInstallRoot}/bin
export GO111MODULE=on
make static-dist bin
""")
            }
          }
        }
      }
    }
  }
}