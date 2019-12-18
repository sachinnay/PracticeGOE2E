

##generate coverage file and test report file for sonrqube 

go test ./... --cover -coverprofile=coverage.out -json | tee test-report.out

##Run sonar -Scanner 
/home/sachin/sonar-scanner-3.2.0.1227-linux/bin/sonar-scanner -X \
 -Dproject.settings=sonar-project.propertie