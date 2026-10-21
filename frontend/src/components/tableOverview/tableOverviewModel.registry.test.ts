import { describe, expect, it } from 'vitest';

import { resolveSqlDialect } from '../../utils/sqlDialectCore';
import { getMetadataDialect, usesTableListOverview } from './tableOverviewModel';

describe('table overview for command-language registry types', () => {
  it('lists tables through DBGetTables instead of information_schema', () => {
    for (const type of ['zookeeper', 'etcd', 'weaviate', 'influxdb']) {
      expect(usesTableListOverview(resolveSqlDialect(type), type), type).toBe(true);
    }
    // 声明了表状态查询或借用 SQL 方言的描述表类型继续走 SQL。
    for (const type of ['questdb', 'tidb', 'cockroachdb', 'timescaledb']) {
      expect(usesTableListOverview(resolveSqlDialect(type), type), type).toBe(false);
    }
    expect(usesTableListOverview('sqlite', 'sqlite')).toBe(true);
    expect(usesTableListOverview('mysql', 'mysql')).toBe(false);
  });

  it('uses the table list for MongoDB, which has no information_schema', () => {
    expect(usesTableListOverview('mongodb', 'mongodb')).toBe(true);
  });

  it('uses the table list for vector, time-series and search engines without information_schema', () => {
    // chroma / qdrant / iotdb 没有描述文件；elasticsearch 与 opensearch 的元数据方言都归一为
    // elasticsearch，五个连接类型因此共用同一条表名列表路径。这里走 getMetadataDialect 而不是
    // 直接传方言，保证归一链路本身也在覆盖范围内。
    for (const type of ['chroma', 'qdrant', 'iotdb', 'elasticsearch', 'opensearch']) {
      const dialect = getMetadataDialect(type);
      expect(usesTableListOverview(dialect, type), `${type} -> ${dialect}`).toBe(true);
    }
  });
});
