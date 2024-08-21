#!/bin/bash

# Number of requests per second
rate=5

# Duration in seconds
duration=300

# Total number of requests
total_requests=$(($rate * $duration))

# Time to sleep between requests to achieve the desired rate
interval=$(echo "scale=2; 1 / $rate" | bc)

# Execute the requests
for ((i=1; i<=total_requests; i++))
do
    ../bin/serverledge-cli invoke -H 192.168.1.245 -P 1324 -f func -p 'a:2' -p 'b:3' &
    # Sleep to control the rate of requests
    sleep $interval
done

# Wait for all background processes to complete
wait
