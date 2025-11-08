# goInception

<!-- [![travis-ci](https://img.shields.io/travis/hanchuanchuan/goInception.svg)](https://travis-ci.org/hanchuanchuan/goInception) -->
[![CircleCI Status](https://circleci.com/gh/hanchuanchuan/goInception.svg?style=shield)](https://circleci.com/gh/hanchuanchuan/goInception)
[![GitHub release](https://img.shields.io/github/release-pre/hanchuanchuan/goInception.svg?style=brightgreen)](https://github.com/hanchuanchuan/goInception/releases)
[![codecov](https://codecov.io/gh/hanchuanchuan/goInception/branch/master/graph/badge.svg)](https://codecov.io/gh/hanchuanchuan/goInception)
[![](https://img.shields.io/badge/go-1.22.1-brightgreen.svg)](https://golang.org/dl/)
[![TiDB](https://img.shields.io/badge/TiDB-v2.1.1-brightgreen.svg)](https://github.com/pingcap/tidb)
![](https://img.shields.io/github/downloads/hanchuanchuan/goInception/total.svg)
![](https://img.shields.io/github/license/hanchuanchuan/goInception.svg)


**[[English]](README.md)**
**[[Chinese]](README.zh.md)**


goInception is a MySQL maintenance tool, which can be used to review, implement, backup, and generate SQL statements for rollback. It parses SQL syntax and returns the result of the review based on custom rules.

> Compatibility: goInception works with MySQL 5.6+ and has been verified against MySQL 8.4 (including the default `caching_sha2_password` authentication).

**Documentation:**
**[[Document]](https://hanchuanchuan.github.io/goInception/)**
**[[中文文档]](https://hanchuanchuan.github.io/goInception/zh/)**

**[[Changelog]](https://hanchuanchuan.github.io/goInception/changelog.html)**


----

### Quick start


#### Binary

[goInception Download](https://github.com/hanchuanchuan/goInception/releases)


#### Docker Image
```
docker pull hanchuanchuan/goinception
```


#### Source code compilation

***go version 1.14+ (go mod)***

```bash
git clone https://github.com/hanchuanchuan/goInception.git
cd goInception
cp config/config.toml.default config.toml
# edit config.toml as needed (listen host/port, security, etc.)
go build -o goInception tidb-server/main.go

./goInception -config=config/config.toml
```

----

#### Associated SQL audit platform

* [Archery](https://github.com/hhyo/Archery) `Query support (MySQL/MsSQL/Redis/PostgreSQL), MySQL optimization (SQLAdvisor|SOAR|SQLTuning), slow log management, table structure comparison, session management, Alibaba Cloud RDS management, etc.`


#### Testing & verification

- **Authentication** – use both the MySQL CLI and go-sql-driver clients to connect to goInception with `caching_sha2_password` (default in MySQL 8.4) and `mysql_native_password`, ensuring Auth Switch handshakes succeed.
- **Remote SQL** – submit control-comment blocks (`/*--...*/inception_magic_start; ...`) via CLI/Go helper to exercise DDL/DML against a MySQL 8.4 server, for example:

  ```bash
  mysql --comments -h127.0.0.1 -P4000 <<'SQL'
  /*--user=root;--password=***;--host=your.mysql.host;--port=3306;--check=1;--execute=0;*/inception_magic_start;
  use test;
  create table t1(id int primary key);
  inception_magic_commit;
  SQL
  ```

- **pt-online-schema-change** – verified against MySQL 8.4 using `--recursion-method=none`, confirming caching_sha2 authentication works.
- **gh-ost** – current release (as of 2024) still issues `SHOW SLAVE STATUS`, which MySQL 8.4 rejects (`SHOW REPLICA STATUS` is required). Once upstream fixes or a compatible build is available, rerun the OSC validation.

#### Acknowledgments
    GoInception reconstructs from the Inception which is a well-known MySQL auditing tool and uses TiDB SQL parser.

- [Inception](https://github.com/hanchuanchuan/inception)
- [TiDB](https://github.com/pingcap/tidb)

#### Sponsorship and support
- [Sponsorship and support](https://hanchuanchuan.github.io/goInception/support.html)

#### Contact

QQ group talk: **499262190**
e-mail: `chuanchuanhan@gmail.com`

### Contributing

Welcome and thank you very much for your contribution. For the process of submitting PR, please refer to [CONTRIBUTING.md](CONTRIBUTING.md)。


## Contributors

### Code Contributors

This project exists thanks to all the people who contribute. [[Contribute](CONTRIBUTING.md)].
<a href="https://github.com/hanchuanchuan/goInception/graphs/contributors"><img src="https://opencollective.com/goInception/contributors.svg?width=890&button=false" /></a>

### Financial Contributors

Become a financial contributor and help us sustain our community. [[Contribute](https://opencollective.com/goInception/contribute)]

#### Individuals

<a href="https://opencollective.com/goInception"><img src="https://opencollective.com/goInception/individuals.svg?width=890"></a>

#### Organizations

Support this project with your organization. Your logo will show up here with a link to your website. [[Contribute](https://opencollective.com/goInception/contribute)]

<a href="https://opencollective.com/goInception/organization/0/website"><img src="https://opencollective.com/goInception/organization/0/avatar.svg"></a>
<a href="https://opencollective.com/goInception/organization/1/website"><img src="https://opencollective.com/goInception/organization/1/avatar.svg"></a>
<a href="https://opencollective.com/goInception/organization/2/website"><img src="https://opencollective.com/goInception/organization/2/avatar.svg"></a>
<a href="https://opencollective.com/goInception/organization/3/website"><img src="https://opencollective.com/goInception/organization/3/avatar.svg"></a>
<a href="https://opencollective.com/goInception/organization/4/website"><img src="https://opencollective.com/goInception/organization/4/avatar.svg"></a>
<a href="https://opencollective.com/goInception/organization/5/website"><img src="https://opencollective.com/goInception/organization/5/avatar.svg"></a>
<a href="https://opencollective.com/goInception/organization/6/website"><img src="https://opencollective.com/goInception/organization/6/avatar.svg"></a>
<a href="https://opencollective.com/goInception/organization/7/website"><img src="https://opencollective.com/goInception/organization/7/avatar.svg"></a>
<a href="https://opencollective.com/goInception/organization/8/website"><img src="https://opencollective.com/goInception/organization/8/avatar.svg"></a>
<a href="https://opencollective.com/goInception/organization/9/website"><img src="https://opencollective.com/goInception/organization/9/avatar.svg"></a>
