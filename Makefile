#
#    Onix Kube - Copyright (c) 2019 by www.gatblau.org
#
#    Licensed under the Apache License, Version 2.0 (the "License");
#    you may not use this file except in compliance with the License.
#    You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
#    Unless required by applicable law or agreed to in writing, software distributed under
#    the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
#    either express or implied.
#    See the License for the specific language governing permissions and limitations under the License.
#
#    Contributors to this project, hereby assign copyright in this code to the project,
#    to be licensed under the same terms as the rest of the code.
#

# the name of the container registry repository
REPO_NAME=gatblau

# the name of the ox-kube binary file
APP_NAME=onix-snapshot

# the name of the go command to use to build the binary
GO_CMD = go

# the version of the application
APP_VER = v0.0.2

# get old images that are left without a name from new image builds (i.e. dangling images)
DANGLING_IMS = $(shell docker images -f dangling=true -q)

# produce a new version tag
version:
	sh version.sh $(APP_VER)

# build the ox-kube docker image
docker-image:
	$(MAKE) version
	docker build -t $(REPO_NAME)/$(APP_NAME):$(shell cat src/main/resources/version) .
	docker tag $(REPO_NAME)/$(APP_NAME):$(shell cat src/main/resources/version) $(REPO_NAME)/$(APP_NAME):latest

docker-push:
	docker push $(REPO_NAME)/$(APP_NAME)-snapshot:$(shell cat src/main/resources/version)
	docker push $(REPO_NAME)/$(APP_NAME)-snapshot:latest

# deletes dangling images
docker-clean:
	docker rmi $(DANGLING_IMS) -f