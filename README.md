
```
link
├─ .gitignore
├─ EnvKey
├─ README.md
├─ cmd
│  └─ main.go
├─ config
│  ├─ config.go
│  ├─ di.go
│  └─ init.go
├─ go.mod
├─ go.sum
├─ infrastructure
│  └─ persistence
│     ├─ auth_persistence_redis.go
│     └─ user_persistence_pg.go
├─ internal
│  ├─ auth
│  │  ├─ entity
│  │  │  └─ token_entity.go
│  │  ├─ repository
│  │  │  └─ auth_repository.go
│  │  └─ usecase
│  │     └─ auth_usecase.go
│  ├─ group
│  │  └─ entity
│  │     └─ group_entity.go
│  └─ user
│     ├─ entity
│     │  └─ user_entity.go
│     ├─ repository
│     │  └─ user_repository.go
│     └─ usecase
│        └─ user_usecase.go
├─ pkg
│  ├─ dto
│  │  ├─ auth
│  │  │  ├─ req
│  │  │  │  └─ auth_req.go
│  │  │  └─ res
│  │  │     └─ auth.res.go
│  │  └─ user
│  │     ├─ req
│  │     │  └─ user_req.go
│  │     └─ res
│  │        └─ user_res.go
│  ├─ http
│  │  ├─ auth_handler.go
│  │  └─ user_handler.go
│  ├─ interceptor
│  │  ├─ error_handler.go
│  │  ├─ response.go
│  │  └─ token_interceptor.go
│  └─ util
│     ├─ jwt.go
│     └─ password.go
└─ tmp
   └─ main

```