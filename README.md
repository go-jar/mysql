# 思路
- query_builder.go：构建 SQL 语句
- client.go：执行 SQL 语句
- dao.go：以上两者的结合
- config.go：配置 client
- pool.go：client 池
- orm.go：直接将对象插入 db，或通过对象更新、删除、查询 db。成员包括 dao、pool。
# 参考
- https://www.jianshu.com/p/a47d9b6353d6
- https://github.com/goinbox/mysql
