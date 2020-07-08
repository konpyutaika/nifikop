#!/bin/bash
if [ -d sonar-scanner-3.3.0.1492-linux ]
then
    echo "Sonar Scanner already cached"
else
    curl -o scanner.zip 'https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-3.3.0.1492-linux.zip'
    mkdir sonar-scanner
    unzip scanner.zip
    chmod +x sonar-scanner-3.3.0.1492-linux/bin/sonar-scanner
fi
