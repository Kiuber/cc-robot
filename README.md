# Cryptocurrency Robot

### How to use

1. Ensure MySQL Redis ready
1. Execute MySQL init script
1. Fill config info in config folder, eg: API key and secret, MySQL, Redis

- Dev:
   - python manager.py --run
- Prod:
   - python manager.py --build
   - python manager.py --restart_app_container

### TODO

1. [ ] Infra
   1. [x] boot
   1. [x] run cli args
   1. [x] mock API
   1. [x] notification
   1. [ ] log rotate
   1. [ ] multi output writer
   1. [ ] clis
   1. [x] MySQL
   1. [x] Redis
   1. [ ] gracefully exit

1. [x] Runtime
   1. [x] Dockerfile

1. [ ] CEX
   1. [ ] mexc
      1. [x] API
      1. [ ] strategy rules
         1. [x] appear symbol pair, buy and sell
