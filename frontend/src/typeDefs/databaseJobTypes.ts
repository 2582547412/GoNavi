export type DatabaseJobKind = 'oracle-scheduler' | 'dbms-job' | 'dm-agent';

export type DatabaseJobEntry = {
  jobName: string;
  schemaName: string;
  jobKind: DatabaseJobKind;
  enabled?: string;
  state?: string;
  nextRunAt?: string;
};

export type DatabaseJobRow = Record<string, unknown>;

export type DatabaseJobQueryResult = {
  success: boolean;
  data?: unknown;
  message?: string;
};

export type DatabaseJobQueryExecutor = (sql: string) => Promise<DatabaseJobQueryResult>;
