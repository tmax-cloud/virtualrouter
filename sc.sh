go build .
sleep 2
docker image rm 192.168.7.191:5000/virtualrouter:0.0.2 .
docker image rm virtualrouter:0.0.2 .
docker build --network host -t virtualrouter:0.0.2 .
docker tag virtualrouter:0.0.2 192.168.7.191:5000/virtualrouter:0.0.2
docker push 192.168.7.191:5000/virtualrouter:0.0.2
