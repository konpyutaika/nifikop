#----- INFO DEFINISSANT LE PROJET DANS SONAR ----
sonar.projectName=cda_analytics_${APPNAME}
sonar.projectKey=com.orange:cda_analytics-${APPNAME}
sonar.branch=${CI_COMMIT_REF_NAME}
#sonar.organization=${SONAR_ORGANIZATION}
#sonar.host.url=${SONAR_HOST_URL}
#sonar.projectBaseDir=/home/circleci/<< parameters.operatorDir >>
sonar.sources=.
sonar.sources.inclusions="**/**.go"
sonar.exclusions="**/*_test.go,**/vendor/**,**/sonar-scanner-3.3.0.1492-linux/**,**docs/**"
sonar.coverage.exclusions="**/vendor/**,**/test/**,**docs/**"
sonar.tests=.
sonar.language=go
sonar.sourceEncoding=UTF-8
sonar.test.inclusions="**/**_test.go"
sonar.test.exclusions="**/vendor/**"
sonar.go.coverage.reportPaths=coverage.out
sonar.go.tests.reportPaths=test-report.out
sonar.coverage.dtdVerification=false
sonar.log.level=INFO
sonar.links.ci=https://github.com/Orange-OpenSource/nifikop/pipelines
sonar.dynamicAnalysis = reuseReports
sonar.scm.enabled = false

#sonar.working.directory=/tmp/_build/.sonar
