# drone-gcr

[![Build Status](http://beta.drone.io/api/badges/drone-plugins/drone-gcr/status.svg)](http://beta.drone.io/drone-plugins/drone-gcr)
[![Coverage Status](https://aircover.co/badges/drone-plugins/drone-gcr/coverage.svg)](https://aircover.co/drone-plugins/drone-gcr)
[![](https://badge.imagelayers.io/plugins/drone-gcr:latest.svg)](https://imagelayers.io/?images=plugins/drone-gcr:latest 'Get your own badge on imagelayers.io')

Drone plugin to build and publish Docker images to Google Container Registry

## Docker

Build the container using `make`:

```
make deps docker
```

### Example

```sh
docker run -i --privileged -v $(pwd):/drone/src plugins/drone-gcr <<EOF
{
    "repo": {
        "clone_url": "git://github.com/drone/drone",
        "owner": "drone",
        "name": "drone",
        "full_name": "drone/drone"
    },
    "system": {
        "link_url": "https://beta.drone.io"
    },
    "build": {
        "number": 22,
        "status": "success",
        "started_at": 1421029603,
        "finished_at": 1421029813,
        "message": "Update the Readme",
        "author": "johnsmith",
        "author_email": "john.smith@gmail.com"
        "event": "push",
        "branch": "master",
        "commit": "436b7a6e2abaddfd35740527353e78a227ddcb2c",
        "ref": "refs/heads/master"
    },
    "workspace": {
        "root": "/drone/src",
        "path": "/drone/src/github.com/drone/drone"
    },
    "vargs": {
        "username": "kevinbacon",
        "password": "pa$$word",
        "email": "foo@bar.com",
        "repo": "foo/bar",
        "storage_driver": "aufs"
    }
}
EOF
```
