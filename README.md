# CMD Challenge

This repository contains the code for the site cmdchallenge.com

## Contribute

Have a suggestion for a challenge? Adding a new one is simple!

Before you do please make sure the challenge you are proposing isn't already
covered by an existing challenge and accepting it is not guaranteed!

* Add a new entry to [challenges.yml](https://gitlab.com/jarv/cmdchallenge/blob/master/challenges.yaml).
  * Pick a unique slug name.
  * Type a description.
  * Add directory and supporting files for the challenge in the `var/challenges` dir. A README will automatically be created in the challenge directory based on the description in challenges.yaml.
  * Add an example solution.
  * Add expected output if the command has output that should be verified.
  * Add a test script to `ro_volume/cmdtests/` if tests are needed after the command is run (see other challenge examples).
  * Add your gitlab username or name to the author field.
* Run `make test` to make sure your new challenge works with the example.
* Submit a merge request (like a pull request on github).

## Installation

* Install docker on your machine
* `pipenv shell`
* `pipenv install`

## Testing

* `make test`

Assuming you have docker installed running `make test` will create a new
docker image, load it and run all tests.

* `./bin/test_challenges <test_name>`
Test a single challenge using the currently built docker container or
all challenges (faster than `make test`).

## CI vars

The following CI vars are necessary to run the full pipeline

* `CA_PEM_FNAME`: file path to `ca.pem`
* `CA_KEY_FNAME`: file path to `ca-key.pem`
* `GCP_CRED_JSON_FNAME`: file path to gcp json credential
* `AWS_ACCESS_KEY_ID`: Access key for AWS
* `AWS_SECRET_ACCESS_KEY`: Secret key for AWS
* `STATE_S3_BUCKET`: where to store terraform state
* `STATE_S3_KEY`: key for storing state
* `STATE_S3_REGION`: region for deployment


### Terraform vars from env

* `TF_VAR_GCP_CRED_JSON_FNAME=GCP_CRED_JSON_FNAME`
* `TF_VAR_CA_PEM_FNAME=CA_PEM_FNAME`

## Bugs / Suggestions

* Open [a gitlab issue](https://gitlab.com/jarv/cmdchallenge/-/issues).
