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
          withEnv(["http_proxy=http://10.192.116.73:8080/","https_proxy=http://10.192.116.73:8080/"]) {
          configFileProvider([configFile(fileId: 'artifactory-npmrc', targetLocation: '.npmrc')]) {
            sshAsVxPipeline {
              sh(returnStdout: true, script: """
export GOROOT=${goInstallRoot}
export GOPROXY=https://goproxy.io
PATH=\$PATH:${nodeInstallRoot}:${nodeInstallRoot}/bin:${yarnInstallRoot}/bin:${goInstallRoot}/bin
export GOBIN=${goInstallRoot}/bin
export GO111MODULE=on
make bootstrap static-dist bin
""")
            }
          }
          }
        }
      }
    }
  }
}