# 数据库作业对象支持

本功能是**只读对象浏览**：列出作业、查看目录属性与原始执行内容，以及已有步骤和调度。
不自动安装插件、启动调度服务或执行作业，不生成伪造的重建 SQL。
新建、修改、删除、启停和立即执行不属于当前实现。

## 已接入的查询适配器

| 数据源 | 机制 | 入口和范围 | 前提及限制 |
| --- | --- | --- | --- |
| Oracle | DBMS_SCHEDULER、DBMS_JOB | schema 下的作业 | 遵循目录可见性；解析关联程序、调度 |
| 达梦 | SYSJOB 作业代理、兼容 DBMS_JOB | schema 下的作业 | 系统表不可读时尝试同所有者的 USER_*_VIEW；不初始化作业环境 |
| SQL Server | SQL Server Agent | **msdb** 下的作业 | 实例级目录；显示步骤、跳转、重试及调度原始字段；不重复挂入每个业务库 |
| PostgreSQL、TimescaleDB | pg_cron、pgAgent | **插件安装所在数据库**下的作业 | 先探测表是否存在；不把执行目标数据库误当作目录所在库；未安装不显示虚假的空作业组 |
| KingbaseES | KDB_SCHEDULE | 当前数据库下的作业 | 先检测 kdb_schedule；owner 可能为空，因此不套用 Oracle 所有者过滤；关联程序/调度读取失败保留主体并警告 |
| openGauss | PG_JOB、PG_JOB_PROC | 当前数据库下的作业 | dbname 使用 current_database() 过滤；较旧版本缺少命名/enable 字段时降级读取基本字段 |
| GaussDB | PG_JOB、PG_JOB_PROC | 当前数据库下的作业 | 同上，独立来源标识；需要目录读取权限 |
| OceanBase Oracle 模式 | DBA_SCHEDULER_JOBS | schema 下的作业 | 独立 SQL，不能按 Oracle 方言直接复用；旧版不公开执行内容时明确显示部分定义提示 |
| OceanBase MySQL 模式 | oceanbase.DBA_SCHEDULER_JOBS | 所选数据库下的**租户级**作业 | 不按业务库过滤 OWNER；定义按目录返回的 OWNER 和 JOB_NAME 精确查询 |
| Doris | JOBS('type'='insert') | 所选数据库下的**集群级**作业 | Doris 2.1+；需对应目录权限；当前为 INSERT 作业，不包含其他专用后台任务 |
| StarRocks | information_schema.tasks | 当前数据库下的任务，归入作业组 | 包含目录中可见的任务；周期任务需支持 SCHEDULE 的版本（3.3+） |
| MySQL、MariaDB | Event Scheduler | 既有**事件**入口 | 沿用事件列表与定义，不另建重复的作业组；查看事件不等于开启 Event Scheduler |

这些是**代码适配范围，不是所有版本/实库已验收的声明**。数据库版本、插件位置和目录权限仍影响可用性。
数据库未公开的字段不推断、不伪造；权限/版本查询失败与“插件未安装”分开处理。

## 待接入

用户已确认：**Db2 和 Snowflake 先列为待接入驱动**。
仓库尚无这两种连接驱动，不能把 SQL 模板或静态能力开关当作可用支持；后续应先完成连接和查询链路，再接入 Db2 Administrative Task Scheduler / Snowflake Tasks。

## 身份和安全边界

- 作业节点和标签页身份包含连接、数据库、来源、所有者及原生名称/ID。
- ID 型目录保留字符串 ID，不用显示名作为身份；pg_cron 与 pgAgent 的同名同 ID 作业也能区分。
- 数据库/实例级作业挂在数据库根部，不创建假的 schema；定位、刷新和会话恢复使用相同范围。
- 所有新查询均为 SELECT，复用 DBQuery 安全控制和连接配置。
- pgAgent 步骤只读取需要的列，不读取可能包含密码的 `jstconnstr`。
- OceanBase MySQL / StarRocks 名称使用 HEX 比较，避免依赖反斜杠字符串转义模式。
- 显式识别引擎及自定义驱动别名，不因为借用 MySQL/PostgreSQL/Oracle 方言就声明调度能力。

## 验证记录（2026-10-14）

- 新增查询、探测、名称转义、版本降级、部分失败、定义展示、标签恢复和定位测试。
- 18 个相关测试文件、202 项测试通过；不代表完整测试套件通过。
- TypeScript 检查、Vite 生产构建及 Brotli 打包通过；构建仍有分块体积警告。
- 覆盖数据库根部作业与 schema 懒加载共存；刷新重新探测插件，不会在读取作业时自动安装扩展。
- PostgreSQL 实库：只读探测成功，当前 `postgres` 数据库没有 pg_cron / pgAgent 目录；这只证明缺失插件检测，不证明真实作业读取。
- Kingbase 实库：连接建立后的验证被远端关闭，未执行成功元数据查询；未修改连接、TLS 或服务器配置。
- 其他数据库没有对应可用实例，真实作业列表、定义及跨版本验收未运行。
- 浏览器 Mock RPC：PostgreSQL 两种插件的同名作业、SQL Server Agent 步骤/调度、StarRocks 任务定义和分组刷新已走通并检查截图。
- Mock RPC 浏览器与单测不能替代原生 Wails/实库验收。

## 官方依据

- PostgreSQL pg_cron：https://github.com/citusdata/pg_cron
- pgAgent 目录定义：https://github.com/pgadmin-org/pgagent/blob/master/sql/pgagent.sql
- SQL Server Agent：https://learn.microsoft.com/en-us/ssms/agent/sql-server-agent
- Kingbase KDB_SCHEDULE：https://help.kingbase.com.cn/v8/admin/reference/ref-extended-plug-in/kdb_schedule.html
- openGauss PG_JOB / PG_JOB_PROC：openGauss 官方《数据库参考》系统表章节。
- GaussDB PG_JOB / PG_JOB_PROC：华为云官方《GaussDB SQL 参考》系统表章节。
- OceanBase DBA_SCHEDULER_JOBS / DBA_SCHEDULER_PROGRAM：OceanBase 官方文档对应租户模式、版本的系统视图章节。
- 目录字段复核：OceanBase 官方仓库 `src/share/inner_table/ob_inner_table_schema.28101_28150.cpp`、`ob_inner_table_schema.21251_21300.cpp`；openGauss 官方仓库 `src/include/catalog/pg_job.h`、`pg_job_proc.h`。
- Doris JOBS：https://doris.apache.org/docs/3.x/sql-manual/sql-functions/table-valued-functions/jobs/
- StarRocks TASKS：https://docs.starrocks.io/docs/sql-reference/information_schema/tasks/
