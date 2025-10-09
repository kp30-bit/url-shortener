#command to create and deploy new image of app every time
docker build -t your-dockerhub-username/kliplink:latest . && \
docker push your-dockerhub-username/kliplink:latest

Scope of Improvement 

1. Real time analytics update using reno and kafka
2. Include ci/cd pipeline using github actions to deploy after any change in project
3. Add observability
4. Add Cache
