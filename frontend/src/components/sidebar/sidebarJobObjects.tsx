import type { MenuProps } from 'antd';
import { CodeOutlined, ReloadOutlined } from '@ant-design/icons';
import type { SavedConnection, TabData } from '../../types';
import type { DatabaseJobEntry, DatabaseJobKind } from '../../typeDefs/databaseJobTypes';
import { t } from '../../i18n';
import type { SidebarTreeNode } from './sidebarV2TreeNodes';
import type { SidebarNodeMenuContext } from './sidebarNodeMenuHelpers';
import { renderSidebarObjectIcon } from './sidebarObjectIcons';
import { createSidebarObjectGroupBuilder } from './sidebarObjectGroup';
import { buildSidebarCopyObjectNameMenuItem } from './sidebarCopyObjectNameMenu';

type JobNode = {
  key?: unknown;
  dataRef?: {
    id?: string;
    dbName?: string;
    schemaName?: string;
    jobName?: string;
    jobKind?: DatabaseJobKind;
  };
};

export const buildDatabaseJobGroup = (
  conn: SavedConnection & { dbName?: string },
  parentKey: string,
  entries: DatabaseJobEntry[],
  schemaName = '',
): SidebarTreeNode => {
  const children: SidebarTreeNode[] = entries.filter((entry) => !schemaName || entry.schemaName === schemaName).map((entry) => ({
    title: entry.jobKind === 'dbms-job' ? `${entry.jobName} · DBMS_JOB` : entry.jobName,
    key: `${parentKey}-job-${encodeURIComponent(JSON.stringify([entry.jobKind, entry.schemaName, entry.jobName]))}`,
    icon: renderSidebarObjectIcon('job'),
    type: 'db-job',
    isLeaf: true,
    dataRef: { ...conn, ...entry },
  }));
  return createSidebarObjectGroupBuilder(conn)(
    parentKey, 'jobs', t('sidebar.object_group.jobs'), renderSidebarObjectIcon('job'), children, { schemaName },
  );
};

export const openDatabaseJobDefinition = (node: JobNode, addTab: (tab: TabData) => void): void => {
  const { id, dbName, schemaName, jobName, jobKind } = node.dataRef || {};
  if (!id || !jobName || !jobKind) return;
  addTab({
    id: `job-def-${encodeURIComponent(JSON.stringify([id, dbName || '', schemaName || '', jobKind, jobName]))}`,
    title: t('sidebar.tab.job_definition', { name: jobName }),
    type: 'job-def',
    connectionId: id,
    dbName,
    schemaName,
    jobName,
    jobKind,
    sidebarLocateKey: String(node.key || ''),
  });
};

export const buildDatabaseJobMenuItems = (
  node: SidebarTreeNode,
  context: SidebarNodeMenuContext,
): MenuProps['items'] => {
  if (node.type === 'db-job') {
    return [{
      key: 'view-job-def', label: t('sidebar.menu.view_job_definition'), icon: <CodeOutlined />,
      onClick: () => openDatabaseJobDefinition(node, context.addTab),
    }, buildSidebarCopyObjectNameMenuItem(node, context.handleCopyTableName, 'copy-job-name')];
  }
  return [{
    key: 'refresh-jobs', label: t('common.refresh'), icon: <ReloadOutlined />,
    onClick: () => { void context.loadTables(context.getDatabaseNodeRef(node.dataRef, String(node.dataRef?.dbName || '')), { ensureFresh: true }); },
  }];
};
