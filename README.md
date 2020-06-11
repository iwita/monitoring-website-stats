# Monitoring

## What is it?
This is a console program that checks availability and monitors various metrics from websites.
#### Utilities
- Websites and time intervals are user-defined
- Users can keep the console app running and monitor the websites
- Every 10s, display the stats for the past 10 minutes for each website

#### Metrics Supported
The following metrics are calculated for each examined time window
 - Max / Average response time*
 - 90th percentile of response times in the examined time window
 - Availability
 
*Additionally in the case of the 10 minute time window, it calculates the percentage of improvement or decrease of the average response time.
#### Alerting
- When a website availability is below 80% for the past 2 minutes
- When availability resumes for the past 2 minutes
- Alerts remain visible on the page for historical reasons

## Example input file
```yaml
websites:

- url: "https://www.example.com"

interval: 500

- url: "https://www.example1.com"

interval: 1000

- url: "https://www.example2.com/"

interval: 2000

- url: "https://www.example3.com/"

interval: 5000

- url: "https://www.example4.com/"

interval: 10000
```

## Ideas for further application improvement

### 1. Persistence
- Use a timeseries database in order to make the application stateful
- Scalability and Reliability of the application will be improved
- Every minute, displays the stats for the past hour for each website

### 2. Formatted Output
While the logic of the application is created the output could be better formatted using either:
- A well formatted string that will be printed in the CLI, and only the required values will be changed overtime.
- An html template and configure the application to be a served in the web.

### 4. No Bias
In order to avoid bias regarding the response time of various websites, those websites should also get accessed by IPs residing in different continents/timezones.

### 5. Scalability
There exist some issues that may occur, if we try to scale this application.

Generating different goroutines for each website offers great scalability for small amounts of websites. Each goroutine gets served by a different core. When the generated goroutines are more that the available cores of the system, more than one goroutines may compete for the resources of a core. This may result in delays related to cache misses or context switching.

#### Network Bandwidth
When a huge number of websites needs to be monitored, the network bandwidth becomes a bottleneck. The number of requests and responses  will reach the network bandwidth limit. Beyond that point, measured metrics regarding the response times will be inaccurate.

More generally this application could be scaled in a distributed system, where the different websites would be served from different nodes. Later, information to be printed from those distributed nodes could be sent to a **master** node. However the raw metrics retrieved do not need to be sen to the master node, and they could be kept locally.

### 6. Live Input Configuration
Another improvement idea would be to check the input periodically in order to add/remove any newly added/removed websites and change any configured intervals as well. 