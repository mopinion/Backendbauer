git pull origin master
sudo docker stop test
sudo docker rm test
sudo docker stop backendbauer
sudo docker rm backendbauer
sudo docker build -t backendbauer .
sudo docker run -ti -p 81:81 --name test backendbauer
