# oss-contribution-checker

Run this command:
```
go install -mod=vendor ./...  && oss-contribution-checker --account {your github account name}
```
Requirement:
-  need to have `.git-neco.yml` in your home directory or `token.txt` in the current directory which contains a github personal token.

Output:
```
# list of your issues and prs
title: {your issue/pr's title}, year: {created yaer}, repositoryURL: {issue/pr's repository}, needToExclude; {excluded from the summery}
...

Summery:
# of Issues:
...
# of PRs:
...
```

TODO
- リポジトリ別の件数
- PRはマージされたかどうかの情報
- 働きっぷりの可視化
- render json
- summary option(year. project)
