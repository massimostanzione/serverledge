#!/bin/sh
bin/serverledge-cli create -f f1 --memory 512 --src examples/fibonacciNout.py --runtime python310 --handler "fibonacciNout.handler"
bin/serverledge-cli create -f f2 --memory 512 --src examples/fibonacciNout.py --runtime python310 --handler "fibonacciNout.handler"
bin/serverledge-cli create -f f3 --memory 128 --src examples/fibonacciNout.py --runtime python310 --handler "fibonacciNout.handler"
bin/serverledge-cli create -f f4 --memory 1024 --src examples/fibonacciNout.py --runtime python310 --handler "fibonacciNout.handler"
bin/serverledge-cli create -f f5 --memory 256 --src examples/fibonacciNout.py --runtime python310 --handler "fibonacciNout.handler"