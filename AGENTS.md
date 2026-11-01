## Temporal Learn

#### What ?
This project is an customer onboarding for internet wallet (like Alipay, Wechat) support customer can create customer account linked with KYC information, identity, eligibility checks for customer. This microservices include Java/Go services with Temporal server.

#### Techstack
This repository is an mono-repository contains backend services implement using Golang and Java
- Golang techstack: go 1.27.1, go/chi framework
- Java techstack: Java 25, Spring Boot 4
- Database : PostgreSQL 18.6 (via docker)
- Durable execution: temporal server via docker

#### How
- Run Golang app with Go installed version 1.27.1
- Run Java Spring Boot 4 app with Gradle wrapper (gradlew) - groovy with installed Java 25 and gradle (command).
