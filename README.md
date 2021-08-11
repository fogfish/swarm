<p align="center">
  <h3 align="center">swarm</h3>
  <p align="center"><strong>...</strong></p>

  <p align="center">
    <!-- Documentation -->
    <a href="http://godoc.org/github.com/fogfish/swarm">
      <img src="https://godoc.org/github.com/fogfish/swarm?status.svg" />
    </a>
    <!-- Build Status  -->
    <a href="https://github.com/fogfish/swarm/actions/">
      <img src="https://github.com/fogfish/swarm/workflows/build/badge.svg" />
    </a>
    <!-- GitHub -->
    <a href="http://github.com/fogfish/swarm">
      <img src="https://img.shields.io/github/last-commit/fogfish/swarm.svg" />
    </a>
    <!-- Coverage -->
    <a href="https://coveralls.io/github/fogfish/swarm?branch=main">
      <img src="https://coveralls.io/repos/github/fogfish/swarm/badge.svg?branch=main" />
    </a>
    <!-- Go Card -->
    <a href="https://goreportcard.com/report/github.com/fogfish/swarm">
      <img src="https://goreportcard.com/badge/github.com/fogfish/swarm" />
    </a>
    <!-- Maintainability -->
    <a href="https://codeclimate.com/github/fogfish/swarm/maintainability">
      <img src="https://api.codeclimate.com/v1/badges/6d525662ecccc2b9ff04/maintainability" />
    </a>
  </p>
</p>

---

TODO

Supported queues:
* In-memory ephemeral
* AWS SQS
* AWS EventBridge


How to benchmark I/O:

```
cd queue/sqs
go test -run=^$ -bench=.
```
