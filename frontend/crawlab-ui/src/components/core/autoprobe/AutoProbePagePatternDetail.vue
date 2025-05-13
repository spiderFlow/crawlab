<script setup lang="tsx">
import { computed } from 'vue';
import { useStore } from 'vuex';
import {
  ClNavLink,
  ClIcon,
  ClAutoProbeSelector,
} from '@/components';
import { translate } from '@/utils';

const props = defineProps<{
  pagePattern: AutoProbeNavItem;
}>();

const t = translate;

// store
const store = useStore();
const { autoprobe: state } = store.state as RootStoreState;
const form = computed<AutoProbe>(() => state.form);

const tableColumns = computed<TableColumns<AutoProbeNavItem>>(() => {
  return [
    {
      key: 'name',
      label: t('components.autoprobe.pagePattern.name'),
      width: '200',
      value: (row: AutoProbeNavItem) => {
        let icon: Icon;
        switch (row.type) {
          case 'field':
            icon = ['fa', 'tag'];
            break;
          case 'list':
            icon = ['fa', 'list'];
            break;
          case 'pagination':
            icon = ['fa', 'ellipsis-h'];
            break;
          default:
            icon = ['fa', 'question'];
        }
        return (
          <ClNavLink onClick={() => {}}>
            <span style={{ marginRight: '5px' }}>
              <ClIcon icon={icon} />
            </span>
            {row.label}
          </ClNavLink>
        );
      },
    },
    {
      key: 'type',
      label: t('components.autoprobe.pagePattern.type'),
      width: '100',
      value: (row: AutoProbeNavItem) => {
        return t(`components.autoprobe.pagePattern.types.${row.type}`);
      },
    },
    {
      key: 'fields',
      label: t('components.autoprobe.pagePattern.fieldCount'),
      width: '80',
      value: (row: AutoProbeNavItem) => {
        if (row.type === 'list') {
          return <ClNavLink label={row.fieldCount} onClick={() => {}} />;
        }
      },
    },
    {
      key: 'selector',
      label: t('components.autoprobe.pagePattern.selector'),
      width: 'auto',
      minWidth: '300',
      value: (row: AutoProbeNavItem) => {
        switch (row.type) {
          case 'field':
            return <ClAutoProbeSelector type="field" rule={row.field} />;
          case 'pagination':
            return <ClAutoProbeSelector type="pagination" rule={row.pagination} />;
        }
      },
    },
  ] as TableColumns<AutoProbeNavItem>;
});
const tableData = computed<TableData<AutoProbeNavItem>>(() => {
  const { pagePattern } = props;
  if (!pagePattern?.children?.length) {
    return [];
  }
  return pagePattern.children.map((item: AutoProbeNavItem) => {
    switch (item.type) {
      case 'list':
        return {
          name: item.name,
          label: item.name,
          type: item.type,
          fieldCount: item.children?.length || 0,
        };
      case 'field':
        return {
          name: item.name,
          label: item.name,
          type: item.type,
          field: item.field,
        };
      case 'pagination':
        return {
          name: item.name,
          label: item.name,
          type: item.type,
          pagination: item.pagination,
        };
      default:
        return {
          name: item.name,
          label: item.name,
          type: item.type,
        };
    }
  }) as TableData<AutoProbeNavItem>;
});

defineOptions({ name: 'ClAutoProbePagePatternDetail' });
</script>

<template>
  <div class="cl-autoprobe-page-pattern-detail">
    <cl-table :columns="tableColumns" :data="tableData" embedded hide-footer />
  </div>
</template>

<style scoped>
.cl-autoprobe-page-pattern-detail {
  .header {
    margin-bottom: 16px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--el-border-color-light);

    h3 {
      margin: 0;
      font-size: 16px;
      font-weight: 500;
    }
  }

  .content {
    width: 100%;

    .stats-section {
      margin-top: 20px;

      h4 {
        margin: 0 0 12px 0;
        font-size: 14px;
        font-weight: 500;
      }

      .stat-card {
        background-color: var(--el-fill-color-light);
        border-radius: 4px;
        padding: 16px;
        text-align: center;
        height: 100%;

        .stat-value {
          font-size: 24px;
          font-weight: 500;
          margin-bottom: 8px;
        }

        .stat-label {
          color: var(--el-text-color-secondary);
          font-size: 14px;
        }
      }
    }

    .page-data-section {
      margin-top: 20px;

      h4 {
        margin: 0 0 12px 0;
        font-size: 14px;
        font-weight: 500;
      }
    }

    .table {
      width: 100%;

      .el-table__inner-wrapper {
        position: relative;
        overflow: unset;
      }

      .el-table__header-wrapper {
        position: sticky;
        top: 0;
        z-index: 1;
      }

      .table-footer {
        padding: 8px 12px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        border-top: 1px solid var(--el-border-color);

        .total-items {
          font-size: 12px;
          color: var(--el-text-color-secondary);
        }
      }
    }

    .json-value {
      margin: 0;
      white-space: pre-wrap;
      word-break: break-word;
      font-family: monospace;
      font-size: 12px;
      max-height: 150px;
      overflow-y: auto;
    }
  }

  .not-found {
    color: var(--el-text-color-secondary);
    font-style: italic;
    text-align: center;
    padding: 20px;
  }
}
</style>
