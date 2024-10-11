# 代码规范
[Uber Go语言规范](https://github.com/xxjwxc/uber_go_guide_cn)
# Git提交规范
## 基本提交信息格式
```
<type>: <subject>
```
## 提交类型 \<type\>
- feat: 新功能（feature）
- fix: 修复bug
- docs: 新增或修改文档
- style: 不影响代码含义的修改，空格、格式化、缺失符号
- refactor: 非bug修复、非新增特性的重构
- test: 新增或修改测试
- chore: 构建过程或辅助工具的变动

## Git分支
- main分支 (加锁，用于发布打release)
- dev分支 (开发分支，用于日常开发)

后续根据需求可新增分支
## 开发及提交流程
```bash
git checkout dev
git pull --rebase
git checkout -b dev-feat-xxx
# 做开发
...
# 开发完成
git add -A
# 一般提交
git commit -m "<type>: <subject>"
# 特定issue提交 [#xxx]为对应issue的id
git commit -m "<type>: [#xxx]<subject>" 
git checkout dev
git pull --rebase
git rebase dev-feat-xxx
#或者
git cherry-pick dev-feat-xxx
# 第一次推送
git push -u origin dev
# 非第一次推送
git push
```