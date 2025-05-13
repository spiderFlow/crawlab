<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue';
import { useStore } from 'vuex';
import { plainClone, translate, selectElement } from '@/utils';
import { ElTree, ElInput, ElScrollbar } from 'element-plus';
import type { FilterNodeMethodFunction } from 'element-plus/es/components/tree/src/tree.type';
import { debounce } from 'lodash';

// i18n
const t = translate;

// store
const store = useStore();
const { autoprobe: state } = store.state as RootStoreState;

// form data
const form = computed<AutoProbe>(() => state.form);

const pageFields = computed(() => form.value?.page_pattern?.fields);
const pageLists = computed(() => form.value?.page_pattern?.lists);
const pagePagination = computed(() => form.value?.page_pattern?.pagination);
const pageData = computed(() => form.value?.page_data || {});

const treeRef = ref<InstanceType<typeof ElTree>>();
const searchKeyword = ref('');
const showSearch = ref(false);
const activeNavItem = ref<AutoProbeNavItem>();
const defaultExpandedKeys = ref<string[]>([]);

// Helper function to recursively process list items
const processListItem = (list: ListRule): AutoProbeNavItem => {
  const children: AutoProbeNavItem[] = [];

  // Add fields directly if they exist
  if (list.item_pattern?.fields && list.item_pattern.fields.length > 0) {
    list.item_pattern.fields.forEach((field: FieldRule) => {
      children.push({
        id: `${list.name}-${field.name}`,
        label: field.name,
        name: field.name,
        icon: ['fa', 'tag'],
        type: 'field',
        field,
      });
    });
  }

  // Recursively process nested lists if they exist
  if (list.item_pattern?.lists && list.item_pattern.lists.length > 0) {
    list.item_pattern.lists.forEach((nestedList: ListRule) => {
      children.push(processListItem(nestedList));
    });
  }

  return {
    id: list.name,
    label: `${list.name} (${children.length})`,
    name: list.name,
    type: 'list',
    icon: ['fa', 'list'],
    children,
  } as AutoProbeNavItem;
};

const computedTreeItems = computed<AutoProbeNavItem[]>(() => {
  if (!form.value?.page_pattern) return [];
  const children: AutoProbeNavItem[] = [];

  // Add fields directly if they exist
  if (pageFields.value) {
    pageFields.value.forEach(field => {
      children.push({
        id: field.name,
        label: field.name,
        name: field.name,
        icon: ['fa', 'tag'],
        type: 'field',
        field,
      });
    });
  }

  // Add lists directly if they exist
  if (pageLists.value) {
    pageLists.value.forEach(list => {
      children.push(processListItem(list));
    });
  }

  // Add pagination if it exists
  if (pagePagination.value) {
    children.push({
      id: 'pagination',
      label: t('components.autoprobe.navItems.pagination'),
      name: t('components.autoprobe.navItems.pagination'),
      type: 'pagination',
      icon: ['fa', 'ellipsis-h'],
      pagination: pagePagination.value,
    });
  }

  return [
    {
      id: 'page',
      label: `${form.value.page_pattern.name} (${children.length})`,
      name: form.value.page_pattern.name,
      type: 'page_pattern',
      icon: ['fa', 'network-wired'],
      children,
    },
  ];
});
const treeItems = ref<AutoProbeNavItem[]>([]);
watch(
  () => state.form,
  () => {
    treeItems.value = plainClone(computedTreeItems.value);
  }
);

// Function to get field data from page data
const getFieldData = (fieldName: string) => {
  if (!pageData.value) return undefined;
  // Use type assertion to treat pageData as a record with string keys
  return (pageData.value as Record<string, any>)[fieldName];
};

// Function to get data for a specific navigation item
const getNavItemData = (item: AutoProbeNavItem) => {
  if (!pageData.value) return undefined;

  // Convert to Record<string, any> to handle dynamic property access
  const data = pageData.value as Record<string, any>;

  switch (item.type) {
    case 'field':
      // For fields, extract from page data by field name
      return getFieldData(item.id);
    case 'list':
      // For lists, get the corresponding array in page data
      return data[item.id];
    case 'pagination':
      // For pagination, extract pagination-related data
      return data.pagination;
    case 'page_pattern':
      // Return all page data for page pattern
      return pageData.value;
    default:
      return undefined;
  }
};

// Context menu functionality
const activeContextMenuNavItem = ref<AutoProbeNavItem>();
const contextMenuVisibleMap = ref<Record<string, boolean>>({});
const isContextMenuVisible = (id: string) => {
  if (!contextMenuItems.value?.length) return false;
  if (!activeContextMenuNavItem.value) return false;
  if (activeContextMenuNavItem.value?.id !== id) return false;
  return contextMenuVisibleMap.value[id] || false;
};
const onActionsClick = (item: AutoProbeNavItem) => {
  activeContextMenuNavItem.value = item;
  contextMenuVisibleMap.value[item.id] = true;
};
const onContextMenuHide = (id: string) => {
  activeContextMenuNavItem.value = undefined;
  contextMenuVisibleMap.value[id] = false;
};
const contextMenuItems = computed(() => {
  if (!activeContextMenuNavItem.value) return [];
  const { id, type } = activeContextMenuNavItem.value;
  if (!contextMenuVisibleMap.value[id]) return [];
  switch (type) {
    case 'field':
      return [
        {
          title: t('common.actions.view'),
          icon: ['fa', 'eye'],
          action: () =>
            selectNode(activeContextMenuNavItem.value as AutoProbeNavItem),
        },
      ];
    case 'list':
      return [
        {
          title: t('common.actions.view'),
          icon: ['fa', 'eye'],
          action: () =>
            selectNode(activeContextMenuNavItem.value as AutoProbeNavItem),
        },
      ];
    default:
      return [];
  }
});

// Search functionality
const onSearchFilter: FilterNodeMethodFunction = (value, data) => {
  if (!value) return true;
  return data.label.toLowerCase().includes(value.toLowerCase());
};

const debouncedFilter = debounce(() => {
  treeRef.value?.filter(searchKeyword.value);
}, 300);

watch(searchKeyword, debouncedFilter);

// Node selection and expansion tracking
const onNodeClick = async (data: AutoProbeNavItem) => {
  await selectNode(data);
};

const selectNode = async (data: AutoProbeNavItem) => {
  const { id } = data;
  activeNavItem.value = data;

  // Display content in the right panel based on selected node
  // (Implementation would depend on what content you want to show)

  // Highlight current node
  setTimeout(() => {
    treeRef.value?.setCurrentKey(id);
  }, 0);
};

const onNodeExpand = (data: AutoProbeNavItem) => {
  defaultExpandedKeys.value.push(data.id);
};

const onNodeCollapse = (data: AutoProbeNavItem) => {
  const idx = defaultExpandedKeys.value.findIndex(id => id === data.id);
  defaultExpandedKeys.value.splice(idx, 1);
};

const onContextMenuClick = (event: MouseEvent, data: AutoProbeNavItem) => {
  event.stopPropagation();
  activeContextMenuNavItem.value = data;
  contextMenuVisibleMap.value[data.id] = true;
};

const onSearchClick = () => {
  showSearch.value = !showSearch.value;
};

const onRefresh = () => {
  // Refresh data if needed
  treeItems.value = plainClone(computedTreeItems.value);
};

// Sidebar resizing
const widthKey = ref('autoprobe.sidebar.width');
const sidebarRef = ref<HTMLElement | null>(null);

onMounted(() => {
  treeItems.value = plainClone(computedTreeItems.value);
});

defineOptions({ name: 'ClAutoProbeDetailTabPatterns' });
</script>

<template>
  <div class="autoprobe-detail-tab-patterns">
    <div ref="sidebarRef" class="sidebar">
      <cl-resize-handle :target-ref="sidebarRef" :size-key="widthKey" />
      <div class="sidebar-actions">
        <cl-icon :icon="['fa', 'refresh']" @click="onRefresh" />
        <cl-icon
          :class="showSearch ? 'selected' : ''"
          :icon="['fa', 'search']"
          @click="onSearchClick"
        />
      </div>
      <div v-if="showSearch" class="sidebar-search">
        <el-input
          v-model="searchKeyword"
          :placeholder="t('common.search.placeholder')"
          clearable
          @clear="
            () => {
              searchKeyword = '';
              showSearch = false;
            }
          "
        />
      </div>
      <el-scrollbar>
        <el-tree
          ref="treeRef"
          node-key="id"
          :data="treeItems"
          :filter-node-method="onSearchFilter"
          :expand-on-click-node="false"
          :default-expanded-keys="defaultExpandedKeys"
          highlight-current
          @node-click="onNodeClick"
          @node-contextmenu="onContextMenuClick"
          @node-expand="onNodeExpand"
          @node-collapse="onNodeCollapse"
        >
          <template #default="{ data }">
            <cl-context-menu
              :visible="isContextMenuVisible(data.id)"
              :style="{ flex: 1, paddingRight: '5px' }"
            >
              <template #reference>
                <div class="node-wrapper" :title="data.label">
                  <span class="icon-wrapper">
                    <cl-icon
                      v-if="data.loading"
                      :icon="['fa', 'spinner']"
                      spinning
                    />
                    <cl-icon v-else :icon="data.icon || ['fa', 'folder']" />
                  </span>
                  <span class="label">
                    {{ data.label }}
                  </span>
                </div>
                <div class="actions">
                  <cl-icon
                    class="more"
                    :icon="['fa', 'ellipsis']"
                    @click.stop="onActionsClick(data)"
                  />
                </div>
              </template>
              <cl-context-menu-list
                v-if="isContextMenuVisible(data.id)"
                :items="contextMenuItems"
                @hide="onContextMenuHide(data.id)"
              />
            </cl-context-menu>
          </template>
        </el-tree>
      </el-scrollbar>
    </div>
    <div class="content">
      <template v-if="activeNavItem?.type === 'field'">
        <cl-auto-probe-field-detail
          :field="activeNavItem"
          :page-data="getNavItemData(activeNavItem)"
        />
      </template>
      <template v-else-if="activeNavItem?.type === 'list'">
        <cl-auto-probe-list-detail
          :list="activeNavItem"
          :page-data="getNavItemData(activeNavItem)"
        />
      </template>
      <template v-else-if="activeNavItem?.type === 'pagination'">
        <cl-auto-probe-pagination-detail
          :pagination="activeNavItem"
          :page-data="getNavItemData(activeNavItem)"
        />
      </template>
      <template v-else-if="activeNavItem?.type === 'page_pattern'">
        <cl-auto-probe-page-pattern-detail :page-pattern="activeNavItem" />
      </template>
      <div v-else class="placeholder">
        {{
          t('components.autoprobe.patterns.selectItem') ||
          'Select an item to view details'
        }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.autoprobe-detail-tab-patterns {
  height: 100%;
  display: flex;

  .sidebar {
    flex: 0 0 240px;
    height: 100%;
    overflow: hidden;
    border-right: 1px solid var(--el-border-color);
    display: flex;
    flex-direction: column;
    position: relative;

    .sidebar-actions {
      height: 41px;
      flex: 0 0 41px;
      padding: 5px;
      display: flex;
      align-items: center;
      gap: 5px;
      color: var(--cl-primary-color);
      border-bottom: 1px solid var(--el-border-color);

      & > * {
        display: flex;
        align-items: center;
      }

      &:deep(.icon) {
        cursor: pointer;
        padding: 6px;
        font-size: 14px;
        width: 14px;
        height: 14px;
        border-radius: 50%;
      }

      &:deep(.icon.selected),
      &:deep(.icon:hover) {
        background-color: var(--cl-primary-plain-color);
      }
    }

    .sidebar-search {
      height: 38px;
      flex: 0 0 38px;
      border-bottom: 1px solid var(--el-border-color);

      &:deep(.el-input .el-input__wrapper) {
        box-shadow: none;
        border: none;
      }
    }

    .el-tree {
      min-width: fit-content;

      &:deep(.el-tree-node__content:hover .actions .icon) {
        display: flex !important;
      }

      &:deep(.el-tree-node__content) {
        width: 100%;
        position: relative;

        .node-wrapper {
          display: flex;
          align-items: center;
          position: relative;
          width: 100%;

          .icon-wrapper {
            width: 20px;
            display: flex;
          }

          .label {
            flex: 0 0 auto;
          }
        }

        .actions {
          display: flex;
          gap: 5px;
          position: absolute;
          top: 0;
          right: 5px;
          height: 100%;
          align-items: center;

          &:deep(.icon.more) {
            display: none;
          }
        }
      }
    }
  }

  .content {
    flex: 1;
    overflow: auto;

    .detail-panel {
      border: 1px solid var(--el-border-color);
      border-radius: 4px;
      padding: 16px;
    }

    .placeholder {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 100%;
      color: var(--el-text-color-secondary);
      font-style: italic;
    }
  }
}
</style>
