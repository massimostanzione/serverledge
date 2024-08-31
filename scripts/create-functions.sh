#!/bin/sh
#!/bin/sh

HOST=192.168.122.22
PORT=1324

bin/serverledge-cli create -H $HOST -P $PORT -f f1 --memory 512 --src examples/fibonacciNout.py --runtime python310 --handler "fibonacciNout.handler"
bin/serverledge-cli create -H $HOST -P $PORT -f f2 --memory 512 --src examples/fibonacciNout.py --runtime python310 --handler "fibonacciNout.handler"
bin/serverledge-cli create -H $HOST -P $PORT -f f3 --memory 128 --src examples/fibonacciNout.py --runtime python310 --handler "fibonacciNout.handler"
bin/serverledge-cli create -H $HOST -P $PORT -f f4 --memory 1024 --src examples/fibonacciNout.py --runtime python310 --handler "fibonacciNout.handler"
bin/serverledge-cli create -H $HOST -P $PORT -f f5 --memory 256 --src examples/fibonacciNout.py --runtime python310 --handler "fibonacciNout.handler"
