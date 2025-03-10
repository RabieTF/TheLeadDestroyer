# TheLeadDestroyer
This software was developed by Aris Meksaoui and Rabie Toaba for a lab project at Université du Havre Normandie.

TheLeadDestroyer is a powerful tool designed to manage and scale Docker Swarm services dynamically. It leverages the **Docker SDK** to deploy and control multiple containers efficiently.

## Features
- Automatically scales services based on workload.
- Manages containers dynamically using **Docker SDK**.
- Built with **Go** for high performance.
- Deployable as a **Docker container**.

## Running TheLeadDestroyer (Docker)
To run TheLeadDestroyer, you need to:
1. **Ensure Docker is installed and running** on your system.
2. **Grant the container access to the Docker daemon** (`/var/run/docker.sock`).
3. **Provide the required environment variables**.

### Run the Container
```sh
docker run -d \
  --name TheLeadDestroyer \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e MIN_REPLICAS=1 \
  -e MAX_REPLICAS=5 \
  -e THRESHOLD=3 \
  rabietf/theleaddestroyer:latest
```

### Alternative: Use an `.env` File
Instead of passing environment variables manually, create a `.env` file:
```sh
MIN_REPLICAS=1
MAX_REPLICAS=5
THRESHOLD=3
```

Then run:
```sh
docker run -d \
  --name TheLeadDestroyer \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --env-file .env \
  rabietf/theleaddestroyer:latest
```

## Interacting with the service

To interact with the service, you can either use a websocket tool of your choosing (such as websocat) by connecting to ws://host/ws and sending the keyword `client` and then sending the MD5 hashes you want to crack.
Or by using the web app provided [here](https://github.com/RabieTF/DestroyersClient)

## Stopping & Removing the Container
To **stop and remove** the container:
```sh
docker stop TheLeadDestroyer && docker rm TheLeadDestroyer
```

## License
This project is licensed under the **MIT License**.

## Need Help?
If you have any issues, feel free to open an **issue** or **pull request** on GitHub! 🚀

## Github links

Backend app: [https://github.com/RabieTF/TheLeadDestroyer](https://github.com/RabieTF/TheLeadDestroyer)
Web app: [https://github.com/RabieTF/DestroyersClient](https://github.com/RabieTF/DestroyersClient)

