# oss-contribution-checker

Run this command:
```
go install -mod=vendor ./...  && oss-contribution-checker --account {github account name}
```
Requirement:
-  need to have `token.txt` in the current directory which contains a github personal token or pass your github api token to `--token` option.

Output:
- It may contain personal info, so no example is provided here. Check it by yourself:D

TODO
- PRはマージされたかどうかの情報
- 働きっぷりの可視化
- render json
