{%hackmd BJrTq20hE %}
# Setup Github Actions for Golang + Postgres to run automated tests
###### tags: `simplebank`
[Youtube](https://www.youtube.com/watch?v=3mzQRJY1GVE&list=PLy_6D98if3ULEtXtNSY_2qN21VCKgoQAE&index=11)
[Read it on DEV](https://dev.to/techschoolguru/how-to-setup-github-actions-for-go-postgres-to-run-automated-tests-81o)

In this article, we will learn how to setup that process using ***Github Action*** to automatically build and run unit tests for our simple bank project, which is written in Golang and uses PostgreSQL as its main database.


## Continuous integration (CI)
Continuous integration (CI) is one important part of the software development process where a shared code repository is ***continuously changing due to new work of a team member being integrated into it***.

To ensure the high quality of the code and reduce potential errors, each integration is usually ***verified by an automated build and test process***.

## How Github Actions works
Github Action is a service offered by Github that has similar functionality as other CI tools like Jenkins, Travis, or CircleCI.

![](https://i.imgur.com/jUCc3fs.png)

## Workflow
In order to use Github Actions, we must define a workflow. 

Workflow is basically an automated procedure that’s made up of one or more jobs. It can be triggered by 3 different ways:

* By an event that happens on the Github repository
* By setting a repetitive schedule
* Or manually clicking on the run workflow button on the repository UI.
![](https://i.imgur.com/hJ3Brmh.png)

### How to create a workflow

To create a workflow, we just need to add a .yml file to the .github/workflows folder in our repository. For example, this is a simple workflow file ci.yml:

```
name: build-and-test

on:
  push:
    branches: [ master ]
  schedule:
    - cron:  '*/15 * * * *'

jobs:
  build:
    runs-on: ubuntu-latest
```

- We can define how it will be triggered using the `on` keyword.
- In this flow, there's an event that will trigger the workflow whenever
    - a change is pushed to the master branch, and 
    - another scheduled trigger that will run the workflow every 15 minute.

Then we define the list of jobs to run in the jobs section of the workflow yaml file.

### Runner
In order to run the jobs, we must specify a runner for each of them. 

`A runner is simply a server` that listens for available jobs, and it will run only 1 job at a time.

We can use Github hosted runner directly, or specify our own self-hosted runner.

![](https://i.imgur.com/j9afHnb.png)

The runners will run the jobs, then report the their progress, logs, and results back to Github, so we can easily check it on the UI of the repository.

We use the run-on keyword to specify the runner we want to use.
```yaml
jobs:
  build:
    runs-on: ubuntu-latest
```

In this example workflow, we’re using Github’s hosted runner for Ubuntu’s latest version.

### Job
Now let’s talk about Job. A job is a set of steps that will be executed on the same runner.

Normally all jobs in the workflow run in parallel, except when you have some jobs that depend on each other, then they will be run serially.

![](https://i.imgur.com/2U5KGbS.png)
```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Check out code
        uses: actions/checkout@v2
      - name: Build server
        run: ./build_server.sh
  test:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - run: ./test_server.sh
```
In this example, we have 2 jobs:

The first one is build, which has 2 steps: check out code, and build server.
The second job is test, which will run the tests of the application.

#### Specify the dependency between jobs
Here we use the needs keyword to say that the test job depends on the build job, so that it can only be run after our application is successfully built.

This test job only has 1 step that runs the test_server.sh script.

### Step
Steps are individual tasks that run serially, one after another within a job. A step can contain 1 or multiple actions.
![](https://i.imgur.com/vbqBwRV.png)

#### Action
Action is basically a standalone command like the one that run the test_server.sh script that we’ve seen before. If a step contains multiple actions, they will be run serially.

An interesting thing about action is that it can be reused. So if someone has already written a github action that we need, we can actually use it in our workflow.

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Check out code
        uses: actions/checkout@v2
      - name: Build server
        run: ./build_server.sh
```

### Step (2)
Here we use the steps keyword to list out all steps we want to run in our job.

The first step is to check out the code from Github to our runner machine. To do that, we just use the Github actions checkout@v2, which has already been written by the Github action team.

The second step is to build our application server. In this case, we provide our own action, which is simply running the build_server.sh script that we’ve created in the repository.

And that’s it!

### Summary
Before jumping in to coding, let’s do a quick summary:

![](https://i.imgur.com/z5U0Tli.png)
* We can trigger a workflow by 3 ways: event, scheduled, or manually.
* A workflow consists of one or multiple jobs.
* A job is composed of multiple steps.
* Each step can have 1 or more actions.
* All jobs inside a workflow normally run in parallel, unless they depend on each other, then in that case, they run serially.
* Each job will be run separately by a specific runner.
* The runners will report progress, logs, and results of the jobs back to github. And we can check them directly on Github repository’s UI.

## Setup a workflow for Golang and Postgres
Alright, now let’s learn how to setup a real workflow for our Golang application so that it can connect to Postgres, and run all the unit tests that we’ve written in previous lectures whenever new changes are pushed to Github.

### Use a template workflow
![](https://i.imgur.com/7bVk1un.png)

As you can see, a new file go.yml is being created under the folder .github/workflows of our repository with this template:


![](https://i.imgur.com/BX11AsE.png)

Now the Test step has finished, but it failed. We know that because of the red x icon next to it.

### Add Postgres service
Let’s search for github action postgres, and open this official Github Action documentation page about creating Postgres service containers.

![](https://i.imgur.com/IxNpl0T.png)

### Add port mapping to Postgres service
```
    services:
      postgres:
        image: postgres:12
        env:
          POSTGRES_USER: root
          POSTGRES_PASSWORD: secret
          POSTGRES_DB: simple_bank
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
```

### Finish

```yaml
name: ci-test

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

jobs:

  test:
    name: Test
    runs-on: ubuntu-latest


    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_USER: root
          POSTGRES_PASSWORD: secret
          POSTGRES_DB: simple_bank
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:

    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.18
      id: go
    
    - name: Check out code into the Go module directory
      uses: actions/checkout@v3

    - name: Install golang-migrate
      run: |
        curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
        sudo mv migrate /usr/bin/migrate
        which migrate

    - name: Run migrations
      run: make migrateup

    - name: Test
      run: make test
```