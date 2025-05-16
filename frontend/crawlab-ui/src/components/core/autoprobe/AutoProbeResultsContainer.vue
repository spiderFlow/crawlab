<script setup lang="tsx">
import { computed, ref } from 'vue';
import { ClTag } from '@/components';
import { translate, getIconByItemType } from '@/utils';
import { TAB_NAME_RESULTS, TAB_NAME_PREVIEW } from '@/constants';
import { CellStyle } from 'element-plus';

const t = translate;

// Props
const props = defineProps<{
  data?: PageData | PageData[];
  fields?: AutoProbeNavItem[];
  activeFieldName?: string;
}>();

// Refs
const resultsContainerRef = ref<HTMLElement | null>(null);

// States
const activeTabName = ref<string | undefined>(TAB_NAME_RESULTS);

// Computed
const resultsVisible = computed(() => !!activeTabName.value);

const resultsTabItems = computed<NavItem[]>(() => [
  {
    id: TAB_NAME_RESULTS,
    title: t('common.tabs.results'),
  },
  {
    id: TAB_NAME_PREVIEW,
    title: t('common.tabs.preview'),
  },
]);

const tableColumns = computed<TableColumns<PageData>>(() => {
  const { fields } = props;
  if (!fields) return [];
  return fields.map(field => {
    return {
      key: field.name,
      label: field.name,
      minWidth: '200',
      value: (row: PageData) => {
        switch (field.type) {
          case 'list':
            return (
              <ClTag
                icon={getIconByItemType('list')}
                label={field.children?.length || 0}
              />
            );
          default:
            return row[field.name!];
        }
      },
    };
  }) as TableColumns<PageData>;
});

const tableData = computed<TableData<PageData | PageData[]>>(() => {
  const { data } = props;
  if (!data) return [];
  if (Array.isArray(data)) {
    return data;
  }
  return [data];
});

const tableCellStyle: CellStyle<PageData> = ({ column }) => {
  const { activeFieldName } = props;
  if (column.columnKey === activeFieldName) {
    return {
      backgroundColor: 'var(--el-color-primary-light-9)',
    };
  }
  return {};
};

// Methods
const onTabSelect = (id: string) => {
  if (activeTabName.value === id) {
    hideResults();
  } else {
    activeTabName.value = id;
  }
};

const hideResults = () => {
  activeTabName.value = undefined;
};

// Resize handler
const heightKey = 'autoprobe.results.containerHeight';
const onSizeChange = (size: number) => {
  // Emit event to parent to adjust layout
  emit('size-change', size);
};

// Emits
const emit = defineEmits<{
  (e: 'size-change', size: number): void;
}>();

defineOptions({ name: 'ClAutoProbeResultsContainer' });
</script>

<template>
  <div
    ref="resultsContainerRef"
    class="autoprobe-results-container"
    :class="[resultsVisible ? 'results-visible' : '']"
  >
    <cl-resize-handle
      v-if="resultsVisible"
      :target-ref="resultsContainerRef"
      :size-key="heightKey"
      direction="horizontal"
      position="start"
      @size-change="onSizeChange"
    />
    <cl-nav-tabs
      :active-key="activeTabName"
      :items="resultsTabItems"
      @select="onTabSelect"
    >
      <template #extra>
        <div class="results-actions">
          <cl-icon
            v-if="resultsVisible"
            color="var(--cl-info-color)"
            :icon="['fa', 'minus']"
            @click="hideResults"
          />
        </div>
      </template>
    </cl-nav-tabs>
    <div class="results" v-if="activeTabName === TAB_NAME_RESULTS">
      <cl-table
        :key="JSON.stringify(tableColumns)"
        :columns="tableColumns"
        :data="tableData"
        :header-cell-style="tableCellStyle"
        :cell-style="tableCellStyle"
        embedded
        hide-footer
      />
    </div>
    <div class="output" v-else-if="activeTabName === TAB_NAME_PREVIEW"></div>
  </div>
</template>

<style scoped>
.autoprobe-results-container {
  position: relative;
  border-top: 1px solid var(--el-border-color);
  overflow: hidden;

  &.results-visible {
    overflow: auto;
    flex: 0 0 50%;
    height: 50%;
  }

  &:not(.results-visible) {
    flex: 0 0 41px !important;
    height: 41px !important;
  }

  .results-actions {
    display: flex;
    align-items: center;
    padding: 0 10px;

    &:deep(.icon) {
      cursor: pointer;
      padding: 6px;
      font-size: 14px;
      width: 14px;
      height: 14px;
      border-radius: 50%;
    }

    &:deep(.icon:hover) {
      background-color: var(--cl-info-plain-color);
    }
  }

  .results {
    height: calc(100% - 41px);

    &:deep(.table) {
      width: 100%;
      height: 100%;
    }

    &:deep(.table .el-table__inner-wrapper) {
      position: relative;
      overflow: unset;
    }

    &:deep(.table .el-table__header-wrapper) {
      position: sticky;
      top: 0;
    }
  }

  .output {
    padding: 10px;
    height: calc(100% - 41px);
    overflow: auto;
    white-space: pre-wrap;

    pre {
      margin: 0;
      font-size: 14px;
      line-height: 1.5;
      color: var(--cl-text-color);
      white-space: pre-wrap;
    }
  }
}
</style>
