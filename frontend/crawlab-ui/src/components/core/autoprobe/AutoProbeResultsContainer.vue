<script setup lang="tsx">
import { computed, ref } from 'vue';
import { CellStyle } from 'element-plus';
import { ClTag } from '@/components';
import { translate, getIconByItemType } from '@/utils';
import { TAB_NAME_RESULTS, TAB_NAME_PREVIEW } from '@/constants';
import useRequest from '@/services/request';

const t = translate;

// Props
const props = defineProps<{
  data?: PageData | PageData[];
  fields?: AutoProbeNavItem[];
  activeFieldName?: string;
  url?: string;
  activeId?: string;
}>();

// Emits
const emit = defineEmits<{
  (e: 'size-change', size: number): void;
}>();

const { post } = useRequest();

// Refs
const resultsContainerRef = ref<HTMLElement | null>(null);
const iframeRef = ref<HTMLIFrameElement | null>(null);
const iframeLoading = ref(true);
const previewRef = ref<HTMLDivElement | null>(null);
const previewLoading = ref(false);
const pagePreview = ref<PagePreview>();
const overlayRef = ref<HTMLDivElement | null>(null);
const scrollPosition = ref({ top: 0, left: 0 });

// States
const activeTabName = ref<string | undefined>(TAB_NAME_RESULTS);
const resultsVisible = ref(true);

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
const onTabSelect = async (id: string) => {
  activeTabName.value = id;
  if (!resultsVisible.value) {
    resultsVisible.value = true;
  }

  // Reset iframe loading state when switching to preview tab
  if (id === TAB_NAME_PREVIEW) {
    iframeLoading.value = true;
    setTimeout(getPreview, 10);
  }
};

const toggleResults = () => {
  resultsVisible.value = !resultsVisible.value;
  if (!activeTabName.value && resultsVisible.value) {
    activeTabName.value = TAB_NAME_RESULTS;
  }
};

const onIframeLoad = () => {
  iframeLoading.value = false;
  iframeRef.value?.addEventListener('focus', event => {
    console.debug(event);
  });
};

const getPreview = async () => {
  const { activeId } = props;
  const rect = previewRef.value?.getBoundingClientRect();
  const viewport: PageViewPort | undefined = rect
    ? {
        width: rect.width,
        height: rect.height,
      }
    : undefined;
  previewLoading.value = true;
  try {
    const res = await post<any, ResponseWithData<PagePreview>>(
      `/ai/autoprobes/${activeId}/preview`,
      {
        viewport,
      }
    );
    pagePreview.value = res.data;
  } finally {
    previewLoading.value = false;
  }
};

// Resize handler
const heightKey = 'autoprobe.results.containerHeight';
const onSizeChange = (size: number) => {
  // Emit event to parent to adjust layout
  emit('size-change', size);
};

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
            color="var(--cl-info-color)"
            :icon="
              resultsVisible
                ? ['fa', 'window-minimize']
                : ['fa', 'window-maximize']
            "
            @click="toggleResults"
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
    <div
      v-else-if="activeTabName === TAB_NAME_PREVIEW"
      ref="previewRef"
      class="preview"
    >
      <!--      <el-skeleton :rows="15" animated v-if="iframeLoading && url" />-->
      <div v-loading="previewLoading" class="preview-container">
        <div v-if="pagePreview" ref="overlayRef" class="preview-overlay">
          <img class="screenshot" :src="pagePreview.screenshot_base64" />
          <div
            v-for="coord in pagePreview.page_items_coordinates"
            :key="coord.id"
            class="element-mask"
            :style="{
              position: 'absolute',
              left: coord.coordinates?.left + 'px',
              top: coord.coordinates?.top + 'px',
              width: coord.coordinates?.width + 'px',
              height: coord.coordinates?.height + 'px',
            }"
          >
            <el-badge
              type="primary"
              :badge-style="{opacity: 0.5}"
            >
              <template #content>
                <span style="margin-right: 5px">
                  <cl-icon :icon="getIconByItemType('field')" />
                </span>
                {{ coord.id }}
              </template>
            </el-badge>
          </div>
        </div>
      </div>
    </div>
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

  .preview {
    overflow: hidden;
    height: calc(100% - 41px);

    .preview-container {
      position: relative;
      width: 100%;
      height: 100%;
      overflow: auto;
      scrollbar-width: none;

      .preview-overlay {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;

        img.screenshot {
          width: 100%;
        }

        .element-mask {
          border: 1px solid var(--el-color-primary);
          border-radius: 4px;
          pointer-events: none;

          &:hover {
            background-color: var(--cl-primary-color);
            opacity: 0.8;
          }
        }
      }
    }
  }
}
</style>
