git pull origin master
sudo docker build -t backendbauer .
sudo docker stop backendbauer
sudo docker rm backendbauer
sudo docker run -d -p 81:81 --name backendbauer backendbauer
