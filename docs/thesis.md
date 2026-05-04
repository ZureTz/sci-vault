# 基于微服务与智能推荐的实验室科研数据协同管理与发现平台实现

## 摘要

针对高校实验室科研资产分散于个人设备而形成"数据孤岛"、传统知识管理系统缺乏语义理解能力的现实问题，本文设计并实现了实验室科研数据协同管理与发现平台 sci-vault。系统采用 Go + Python 异构微服务架构：基于 Gin 的业务网关 svc-gateway 处理高并发 I/O 与多租户业务，基于 Python 的智能服务 svc-recommender 调用 Gemini 大模型完成元数据萃取与文本向量化，两者通过 gRPC 通信；底层使用 PostgreSQL + pgvector 存储关系型数据与 768 维文档向量，RustFS 提供 S3 兼容对象存储，Redis 缓存任务状态与查询向量。

平台围绕实验室空间（Lab）构建多租户 RBAC 模型，通过异步 Enrichment 流水线实现"上传—解析—摘要—向量化—入库"的自动化知识治理；融合用户点赞、浏览与搜索历史三类行为信号生成画像质心进行个性化推荐；并以 Redis → PostgreSQL → Gemini 三级缓存降低查询嵌入的回源开销。实测显示，CRUD 接口稳定在数十毫秒级；不依赖查询侧嵌入的相似文档推荐处于百毫秒级；而需对用户输入或搜索历史现场嵌入的语义搜索与个性化推荐呈双峰分布——缓存命中约百余毫秒，全 miss 时升至 1–2 秒，这也正是双层向量缓存的设计动因。语义检索相比关键词检索召回率显著提升，推荐结果与用户研究方向相关性良好。本文为中小型实验室低成本、轻量化集成 RAG 与向量检索能力提供了可复用的工程实践方案。

**关键词**：微服务架构；gRPC；语义检索；向量数据库；推荐系统；大语言模型；科研数据管理

# Implementation of a Laboratory Research-Data Collaboration and Discovery Platform Based on Microservices and Intelligent Recommendation

## ABSTRACT

To address the "data silos" caused by lab research assets being dispersed across personal devices, and the lack of semantic understanding in traditional knowledge-management systems, this thesis designs and implements **sci-vault**, a laboratory research-data collaboration and discovery platform. It adopts a heterogeneous Go + Python microservice architecture: a Gin-based gateway, *svc-gateway*, handles high-concurrency I/O and multi-tenant business logic, while a Python service, *svc-recommender*, drives Gemini-LLM metadata extraction and text embedding; the two communicate over gRPC. PostgreSQL with the pgvector extension stores relational data and 768-dimensional document vectors, RustFS provides S3-compatible object storage, and Redis caches task state and query embeddings.

The platform builds a Lab-scoped multi-tenant RBAC model, automates the full "upload → parse → summarize → embed → persist" workflow via an asynchronous enrichment pipeline, and fuses three behavioral signals (likes, views, search history) into a user-profile centroid for personalized recommendation. A Redis → PostgreSQL → Gemini three-tier query-embedding cache amortizes re-embedding cost. Measurements show CRUD APIs respond at tens of milliseconds; the similar-document recommender, which reuses pre-computed vectors, stays in the low hundreds of milliseconds. Endpoints that must embed user input or search history on the fly — semantic search and personalized recommendation — exhibit a bimodal latency profile: roughly 100+ ms on cache hit but 1–2 s on a full miss, which directly motivates the cache design. Semantic retrieval substantially outperforms keyword search in recall, and recommended documents correlate well with users' research interests. This work offers a reusable engineering blueprint for integrating RAG and vector retrieval into small-to-mid laboratories at low cost.

**Keywords**: Microservice Architecture; gRPC; Semantic Retrieval; Vector Database; Recommender System; Large Language Model; Research Data Management

---

# 第 1 章 绪论

## 1.1 研究背景与意义

### 1.1.1 科研数据爆炸式增长与"数据孤岛"困境
- 高校实验室积累的多模态资产（PDF 文献、实验数据、代码、PPT 汇报）现状
- 离散存储介质（个人电脑、微信、移动硬盘、公有云盘）带来的协作障碍
- 新成员入门成本高、知识无法沉淀传承

### 1.1.2 智能化知识发现的迫切需求
- 关键词检索在专业领域的"语义鸿沟"问题
- 从"人找信息"向"信息找人"的范式转变需求
- 大语言模型与向量检索为低成本智能化提供新路径

### 1.1.3 研究意义
- 工程意义：异构微服务架构在中小型场景的最佳实践
- 学术意义：RAG 与轻量化推荐算法在专业垂直场景的工程化验证
- 应用意义：科研资产持久化沉淀与协作效率提升

## 1.2 国内外研究现状

### 1.2.1 知识管理系统（KMS）发展现状
- 国外：Confluence、SharePoint 等成熟产品的功能与架构特征及局限
- 国内：钉钉文档、企业云盘等"数字化存储"方案的局限

### 1.2.2 推荐系统的技术演进
- 协同过滤的冷启动困境与小用户量场景下的失效
- 内容感知（Content-based）与混合推荐的发展
- 基于行为信号融合（点赞 / 浏览 / 搜索历史）的个性化推荐

### 1.2.3 语义检索与 RAG 技术现状
- 词嵌入到上下文嵌入的演进
- pgvector、Milvus、Faiss 等向量数据库技术对比
- RAG 范式与 LLM 驱动的元数据萃取
- 当前缺口：低成本、轻量化集成到中小实验室微服务架构

## 1.3 研究内容
- 异构微服务架构设计与跨语言通信机制
- 基于 Lab 的多租户数据模型与 RBAC 权限体系
- 基于 LLM 的非结构化文档自动化萃取流水线
- 基于 pgvector 的语义检索与个性化推荐引擎
- 三级查询向量缓存与系统性能优化

## 1.4 论文组织结构

---

# 第 2 章 相关技术与理论基础

## 2.1 微服务架构与云原生
- 单体架构 vs 微服务架构的对比
- 服务拆分原则：业务边界与技术异构
- Docker / Docker Compose 容器化与编排

## 2.2 跨语言通信：gRPC 与 Protocol Buffers
- HTTP/REST 与 gRPC 的性能对比
- Protobuf IDL 与服务契约
- 流式 RPC（Server Streaming）在长耗时 LLM 场景的适用性

## 2.3 关系型与向量混合存储
- PostgreSQL 18 的特性
- pgvector 扩展与 HNSW 索引原理
- 余弦距离 `<=>` 与向量召回质量

## 2.4 大语言模型与文本嵌入
- Gemini 模型族概述
- gemini-embedding-001 的语义空间特性
- `RETRIEVAL_QUERY` 与 `RETRIEVAL_DOCUMENT` 任务类型的非对称性
- 嵌入向量在推荐系统中的应用

## 2.5 现代前端技术栈
- SvelteKit 2 与 Svelte 5 Runes 响应式模型
- TailwindCSS v4 原子化样式
- shadcn-svelte / Bits UI 组件体系

## 2.6 对象存储与缓存中间件
- S3 协议与 RustFS 自托管对象存储
- Redis 在异步任务状态、缓存与去重中的角色

---

# 第 3 章 系统需求分析与总体设计

## 3.1 需求分析

### 3.1.1 用户角色与场景
- 实验室管理员、普通成员、未加入用户
- 个人空间与实验室共享空间双场景

### 3.1.2 功能性需求
- 用户管理：注册、登录、个人资料、头像
- 实验室管理：创建、邀请码加入、成员管理、退出/解散
- 文档管理：上传、列表、详情、元数据修改、删除、可见性切换
- 智能化：自动摘要 / 标签萃取、语义搜索、相似文献推荐、个性化推送
- 互动：点赞、浏览历史、搜索历史
- 统计：仪表盘统计指标
- 国际化：中英文切换

### 3.1.3 非功能性需求
- 性能：核心 API 毫秒级响应
- 安全：JWT 鉴权、密码哈希、行级访问控制、SQL 注入防护
- 可扩展：服务可独立横向扩展
- 可维护：分层架构、清晰的服务边界
- 国际化与无障碍：i18n（en、zh-CN）

## 3.2 系统总体架构

### 3.2.1 整体分层
- 前端 SPA 层（SvelteKit）
- 业务网关层（svc-gateway，Go/Gin）
- 智能服务层（svc-recommender，Python）
- 数据存储层（PostgreSQL + pgvector / Redis / RustFS）
- 外部依赖：Gemini API

### 3.2.2 服务间通信
- 浏览器 ↔ svc-gateway：HTTPS REST + JWT
- svc-gateway ↔ svc-recommender：gRPC + Protobuf
- 服务 ↔ 数据层：GORM（Go） / psycopg + pgvector adapter（Python）

### 3.2.3 部署拓扑
- docker-compose 编排所有服务与基础设施
- Nginx 反向代理与静态资源分发

## 3.3 数据库设计

### 3.3.1 实体关系
- `users`：用户基本信息与认证凭据
- `user_profiles`：扩展用户资料
- `labs` / `lab_members`：实验室及多对多成员关系
- `documents`：文档元数据（含可见性、enrich_status、embedding 等）
- `document_views` / `document_likes`：用户互动记录
- `search_histories`：搜索历史
- `query_embeddings`：查询向量持久化缓存（gateway 建表，recommender 读写）

### 3.3.2 关键索引
- `documents` 私有去重唯一索引：`(uploaded_by_user_id, content_sha256) WHERE visibility='private'`
- `documents` Lab 去重唯一索引：`(lab_id, content_sha256) WHERE visibility='lab'`
- `documents.embedding` HNSW 余弦索引
- `query_embeddings` 复合主键 `(query_hash, task_type)` 的成因

### 3.3.3 软删除与一致性
- GORM 软删除（`deleted_at`）在去重与访问控制查询中的过滤约束
- 跨服务表所有权（gateway 建表 / recommender 独占读写）的约束

## 3.4 接口设计

### 3.4.1 REST API 概览
- 路由前缀、URL 参数中间件（`ExtractDocID` / `ExtractLabID`）
- 错误码与 i18n 编码（`service.<resource>.<outcome>`）

### 3.4.2 gRPC 接口契约
- `Health`、`EnrichDocument`（异步）、`TranslateText`（流式）
- `SemanticSearch`、`RecommendSimilar`、`RecommendForUser`

---

# 第 4 章 业务网关 svc-gateway 设计与实现

## 4.1 整体分层与目录结构
- handler / service / repo 三层职责划分
- 错误链：sentinel error → 错误码 → i18n 文案

## 4.2 用户与认证模块
- 注册、登录与密码 bcrypt 哈希
- JWT 签发与中间件鉴权
- 用户资料与头像上传

## 4.3 实验室多租户模块
- 实验室创建、邀请码生成与加入流程
- 成员角色（owner / member）的 RBAC 边界
- 退出与解散的级联处理

## 4.4 文档管理模块
- 多部分上传与 SHA-256 去重
- 文档元数据 CRUD 与 PATCH 部分更新
- 可见性切换（private ↔ lab）的所有权校验
- 软删除与级联清理（互动记录、对象存储）

## 4.5 互动与历史模块
- 浏览节流（15 分钟窗口）与计数器维护
- 点赞 toggle 语义与部分唯一索引
- 搜索历史的 upsert 模型

## 4.6 中间件与安全
- URL 参数提取中间件
- 横向越权防护：调用 svc-recommender 前的 lab 成员关系校验
- SQL 注入防护：参数化查询与排序字段白名单

## 4.7 缓存与异步副作用
- 仪表盘统计 Redis 缓存与失效
- 文档写操作触发的 EnrichDocument gRPC 异步调度
- "best-effort" 副作用的失败容忍策略

## 4.8 与 svc-recommender 的 gRPC 集成
- 客户端封装与超时 / 重试
- 流式翻译接口的转发实现

---

# 第 5 章 智能服务 svc-recommender 设计与实现

## 5.1 服务结构与依赖
- servicer / service / repository / genai 四层结构
- psycopg 与 pgvector 适配器的连接级注册

## 5.2 文档萃取流水线（EnrichDocument）
- 异步 ACK + 后台线程的"先返回后处理"语义
- PDF 文本提取
- Gemini 元数据萃取（标题、摘要、领域标签）
- 768 维 `RETRIEVAL_DOCUMENT` 向量生成
- Redis 任务状态机与 PostgreSQL 终态持久化
- 异构文档支持（Word / PowerPoint / Excel）的扩展点设计 *（待实现，预留章节）*

## 5.3 语义搜索服务（SemanticSearch）
- 查询向量化（`RETRIEVAL_QUERY` 任务类型）
- pgvector 余弦距离召回
- 关键词回退策略
- 访问控制 SQL 片段（私有 + Lab 可见）的复用

## 5.4 个性化推荐服务（RecommendForUser）
- 三类信号采集：点赞、浏览、搜索查询
- 搜索查询使用 `RETRIEVAL_DOCUMENT` 嵌入空间的原因分析
- 用户画像质心计算
- 最近邻召回与去重去自身

## 5.5 相似文档推荐（RecommendSimilar）
- 源文档向量获取
- 同一访问控制作用域内的最近邻
- 与"语义搜索"的实现复用

## 5.6 翻译服务（TranslateText）
- gRPC Server Streaming 模式
- LLM 流式输出的逐 token 转发

## 5.7 三级查询向量缓存
- Redis 字符串 KV 缓存（hex key）
- PostgreSQL `query_embeddings` 复合主键持久层
- Gemini 批量回源
- `resolve_many` 批处理 vs 循环 `resolve` 的 N+1 对比
- 任务类型在 cache key 中的命名空间作用

## 5.8 共享数据结构
- `ScoredDocument` 通用结果类型
- 访问控制 SQL 片段的可组合性

---

# 第 6 章 前端实现

## 6.1 技术栈与工程化
- SvelteKit 2 + Svelte 5 Runes 模式
- TailwindCSS v4 + shadcn-svelte + lucide 图标
- Bun 包管理与 Vite 构建

## 6.2 路由与布局
- `(dashboard)` 路由组与 AppSidebar 注入
- 登录态保护与 401 拦截重定向

## 6.3 状态管理与 Runes
- `.svelte.ts` runes-based store 模式
- `lab.svelte.ts` 与 sidebar 作为 Lab 列表唯一来源
- Effect ID 而非整对象依赖以避免冗余请求
- 路由参数变化下的 `$effect` vs `onMount` 选择

## 6.4 核心页面
- 登录与欢迎页
- 仪表盘（统计指标）
- 实验室列表与设置、成员管理
- 个人 / Lab 文档列表（DataTable 服务端分页）
- 文档详情（萃取信息高亮、相似推荐）
- 语义搜索（自动补全、历史）
- 个性化推荐流
- 个人资料与设置

## 6.5 API 层与错误处理
- Axios 实例与 JWT 拦截器
- `showApiErrors` 与 i18n 错误码联动

## 6.6 国际化
- svelte-i18n 与 en / zh-CN 消息束
- 与后端错误码命名空间对齐

## 6.7 用户体验细节
- 主题切换（mode-watcher 暗色模式）
- 加载骨架屏与异步状态反馈
- 面包屑与返回行为（`afterNavigate`）

---

# 第 7 章 系统测试与性能分析

## 7.1 测试策略
- 单元测试（Go `go test`、Python pytest）覆盖范围
- 集成测试：基于 docker-compose 的端到端流程
- 接口测试：Postman 集合与 gRPC 测试

## 7.2 功能测试
- 用户与实验室核心流程用例
- 文档生命周期与去重 / 软删除场景
- 语义搜索与推荐结果正确性

## 7.3 性能测试
- CRUD 类核心接口响应延迟与 P95 / P99（数十毫秒级）
- 按"是否需要在请求路径上嵌入查询文本"对智能化接口分类测量：
  - **纯库内向量检索**：`RecommendSimilar` 直接读取源文档已沉淀的向量并做最近邻召回，无外部 API 依赖，延迟稳定在百毫秒级
  - **依赖查询侧嵌入**：`SemanticSearch` 与 `RecommendForUser`（需对最近搜索历史做嵌入）呈双峰分布——三级缓存命中约 100+ ms，全 miss 回源 Gemini Embedding API 时延伸至 1–2 s
- 大文件上传吞吐
- gRPC 跨服务调用开销与 LLM 萃取流水线的端到端时延
- pgvector HNSW 检索在不同向量规模下的延迟
- 三级查询向量缓存命中率与冷启动开销节省

## 7.4 安全测试
- JWT 失效 / 篡改场景
- 横向越权（跨 Lab 访问、跨用户私有文档访问）
- SQL 注入与排序字段白名单回归

## 7.5 检索与推荐效果评估
- 关键词检索 vs 语义检索召回率对比
- 主观相关性评估（小样本人工标注）
- `RETRIEVAL_QUERY` / `RETRIEVAL_DOCUMENT` 任务空间混用消融

---

# 第 8 章 总结与展望

## 8.1 工作总结
- 异构微服务架构落地
- 多租户协同与 RBAC 数据安全
- LLM 驱动的自动化知识治理
- 语义检索与三信号融合个性化推荐
- 三级查询向量缓存的工程价值

## 8.2 创新点回顾
- Go + Python "存算分离"的最佳实践
- 向量空间下的语义级资源检索
- 嵌入任务类型非对称性在多信号画像中的处理

## 8.3 不足与展望
- 异构办公文档（Word / PPT / Excel）格式泛化处理 *（待补充：LibreOffice Headless 转换 PDF + Excel 文本序列化）*
- 检索阶段的混合排序（向量 + BM25）
- 知识图谱与实体链接增强
- 模型本地化部署与隐私保护

## 8.4 结语

---

# 参考文献

[1] Newman S. Building Microservices: Designing Fine-Grained Systems[M]. O'Reilly Media, 2021.

[2] gRPC Authors. Introduction to gRPC[EB/OL]. https://grpc.io/docs/what-is-grpc/introduction/.

[3] PostgreSQL Global Development Group. PostgreSQL Documentation[EB/OL]. https://www.postgresql.org/.

[4] Falk K. Practical Recommender Systems[M]. Manning Publications, 2019.

[5] Lewis P, Perez E, Piktus A, et al. Retrieval-Augmented Generation for Knowledge-Intensive NLP Tasks[J]. arXiv:2005.11401, 2021.

[6] Svelte Authors. Svelte — Web development for the rest of us[EB/OL]. https://svelte.dev/.

[7] Donovan A A A, Kernighan B W. The Go Programming Language[M]. Addison-Wesley, 2016.

[8] Malkov Y A, Yashunin D A. Efficient and Robust Approximate Nearest Neighbor Search Using Hierarchical Navigable Small World Graphs[J]. IEEE TPAMI, 2018.

[9] pgvector Authors. pgvector: Open-source vector similarity search for Postgres[EB/OL]. https://github.com/pgvector/pgvector.

[10] Google. Gemini API & Embeddings Documentation[EB/OL]. https://ai.google.dev/.

*（最终引用列表将在论文撰写过程中补充完整。）*

---

# 致谢

*（致谢部分将在论文定稿前撰写。）*
