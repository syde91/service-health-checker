# service-health-checker
A simple concurrent health check monitoring service that monitors the Health(HTTP Status Codes) of the services mentioned in target.csv.

![alt text](https://user-images.githubusercontent.com/80057294/121485936-4538e000-ca03-11eb-9e9d-1ab25a107c08.png)

# Quick Start

To Install, you need to install Go and set your Go workspace first. And execute when the folder is within your GOROOT/src (or) GOPATH.

```
$ make watch
```

## Execute Docker file
```
 docker build -t service-health-check . && docker run  -dp 8080:8080 service-health-check (Change port number as per preference)
 ```
 
# Instructions
1) To add more URLs to your service edit the target.csv file
2) Modify Config file to suit your needs.

# Configuration
Edit config file in /config/config.go.
1) Source: Source file containing all the service URLs
2) MaxConcurrentThreads: Number of concurrent requests that can be made by the application. (Default: 1024)
3) HealthCheckFrequency: Describes frequency of the service health check in seconds. Default: 600 seconds(10 mins)
4) Timeout: The timeout duration for the health check request.

*Note: The application pings all the services in order mentioned in the CSV file and Repings the service only after the all the services are pinged once. To ensure all services are pinged every within the periodic time limit, make sure MaxConcurrentThreads >= (NumberOfServices * Timeout / HealthCheckFrequency)

# CI Pipelines
Pipelines: On every pull request
Pre-submit Checks:
1) Check for lints
2) Check builds
3) Run tests before deployment
4) Check URL format

# Considerations for CD
Since this is a single node application. We need to worry about:
1) Adding a persistance layer under the application, so that we can resume from the place where the application restarted.
2) Eventually we need to break the application to support multi node deployment.
