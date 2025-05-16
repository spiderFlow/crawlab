<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useStore } from 'vuex';
import { cloneDeep } from 'lodash';
import { getIconByExtractType, getIconByItemType, translate } from '@/utils';
import { useAutoProbeDetail } from '@/views';

// i18n
const t = translate;

// store
const store = useStore();
const { autoprobe: state } = store.state as RootStoreState;

const { activeId } = useAutoProbeDetail();

// form data
const form = computed<AutoProbe>(() => state.form);
const pageFields = computed(() => form.value?.page_pattern?.fields);
const pageLists = computed(() => form.value?.page_pattern?.lists);
const pagePagination = computed(() => form.value?.page_pattern?.pagination);
const pageData = computed<PageData>(() => form.value?.page_data || {});
const pageNavItemId = 'page';

// results data and fields based on active item
const resultsDataFields = computed<AutoProbeResults>(() => {
  const rootDataFields: AutoProbeResults = {
    data: pageData.value,
    fields: computedTreeItems.value[0].children?.filter(
      item => item.type !== 'pagination'
    ),
  };

  if (!activeNavItem.value || !pageData.value) {
    return rootDataFields;
  }

  const item = activeNavItem.value;

  if (item.level === 0) {
    return rootDataFields;
  } else if (item.level === 1) {
    if (item.type === 'list') {
      return {
        data: pageData.value[item.name!],
        fields: item.children,
      } as AutoProbeResults;
    }
    return {
      ...rootDataFields,
      activeField: item,
    };
  } else {
    let currentItem = item;
    while (currentItem.parent) {
      const parent = currentItem.parent;
      if (parent.level === 1 && parent.type === 'list') {
        return {
          data: pageData.value[parent.name!],
          fields: parent.children,
          activeField: currentItem,
        } as AutoProbeResults;
      }
      currentItem = currentItem.parent;
    }
    return rootDataFields;
  }
});
const resultsData = computed(() => resultsDataFields.value.data);
const resultsFields = computed(() => resultsDataFields.value.fields);
const resultsActiveField = computed(() => resultsDataFields.value.activeField);

const normalizeItem = (item: AutoProbeNavItem) => {
  const label = item.label ?? `${item.name} (${item.children?.length || 0})`;
  let icon: Icon;
  if (item.type === 'field') {
    const field = item.rule as FieldRule;
    icon = getIconByExtractType(field.extraction_type);
  } else {
    icon = getIconByItemType(item.type);
  }
  return {
    ...item,
    label,
    icon,
  } as AutoProbeNavItem;
};

// Helper function to recursively process list items
const processListItem = (
  list: ListRule,
  parent?: AutoProbeNavItem,
  level: number = 1
): AutoProbeNavItem => {
  const listItem: AutoProbeNavItem = {
    id: list.name,
    name: list.name,
    type: 'list',
    rule: list,
    children: [],
    parent,
    level,
  };

  // Add fields directly if they exist
  if (list.item_pattern?.fields && list.item_pattern.fields.length > 0) {
    list.item_pattern.fields.forEach((field: FieldRule) => {
      listItem.children!.push(
        normalizeItem({
          id: `${list.name}-${field.name}`,
          label: field.name,
          name: field.name,
          type: 'field',
          rule: field,
          parent: listItem,
          level: level + 1,
        })
      );
    });
  }

  // Recursively process nested lists if they exist
  if (list.item_pattern?.lists && list.item_pattern.lists.length > 0) {
    list.item_pattern.lists.forEach((nestedList: ListRule) => {
      listItem.children!.push(processListItem(nestedList, listItem, level + 1));
    });
  }

  return normalizeItem(listItem);
};

// items
const activeNavItem = ref<AutoProbeNavItem>();
const detailNavItem = computed<AutoProbeNavItem | undefined>(() => {
  if (!activeNavItem.value?.type) return;
  switch (activeNavItem.value.type) {
    case 'page_pattern':
    case 'list':
      return activeNavItem.value;
    case 'field':
    case 'pagination':
      return activeNavItem.value.parent;
  }
});
const computedTreeItems = computed<AutoProbeNavItem[]>(() => {
  if (!form.value?.page_pattern) return [];

  const rootItem: AutoProbeNavItem = {
    id: pageNavItemId,
    name: form.value.page_pattern.name,
    type: 'page_pattern',
    children: [],
    level: 0,
  };

  // Add fields directly if they exist
  if (pageFields.value) {
    pageFields.value.forEach(field => {
      rootItem.children!.push(
        normalizeItem({
          id: field.name,
          label: field.name,
          name: field.name,
          type: 'field',
          rule: field,
          parent: rootItem,
          level: 1,
        })
      );
    });
  }

  // Add lists directly if they exist
  if (pageLists.value) {
    pageLists.value.forEach(list => {
      rootItem.children!.push(processListItem(list, rootItem, 1));
    });
  }

  // Add pagination if it exists
  if (pagePagination.value) {
    rootItem.children!.push(
      normalizeItem({
        id: 'pagination',
        label: t('components.autoprobe.navItems.pagination'),
        name: t('components.autoprobe.navItems.pagination'),
        type: 'pagination',
        rule: pagePagination.value,
        parent: rootItem,
        level: 1,
      })
    );
  }

  return [normalizeItem(rootItem)];
});
const treeItems = ref<AutoProbeNavItem[]>([]);
watch(
  () => state.form,
  () => {
    treeItems.value = cloneDeep(computedTreeItems.value);
    if (!activeNavItem.value) {
      activeNavItem.value = treeItems.value[0];
    }
  },
  { immediate: true }
);

watch(activeId, () => {
  // Reset active item when the page changes
  activeNavItem.value = undefined;
});

// ref
const sidebarRef = ref();
const detailContainerRef = ref<HTMLElement | null>(null);

const onNodeSelect = (item: AutoProbeNavItem) => {
  activeNavItem.value = item;
};

const onItemRowClick = (id: string) => {
  const item = sidebarRef.value?.getNode(id);
  if (!item) return;
  activeNavItem.value = sidebarRef.value?.getNode(id);
};

// Handle results container resize
const onSizeChange = (size: number) => {
  if (!detailContainerRef.value) return;
  detailContainerRef.value.style.flex = `0 0 calc(100% - ${size}px)`;
  detailContainerRef.value.style.height = `calc(100% - ${size}px)`;
};

defineOptions({ name: 'ClAutoProbeDetailTabPatterns' });
</script>

<template>
  <div class="autoprobe-detail-tab-patterns">
    <cl-auto-probe-page-patterns-sidebar
      ref="sidebarRef"
      :active-nav-item-id="activeNavItem?.id"
      :tree-items="treeItems"
      :default-expanded-keys="[pageNavItemId]"
      @node-select="onNodeSelect"
    />
    <div class="content">
      <div ref="detailContainerRef" class="detail-container">
        <template v-if="detailNavItem">
          <cl-auto-probe-item-detail
            :item="detailNavItem"
            :active-id="activeNavItem?.id"
            @row-click="onItemRowClick"
          />
        </template>
        <div v-else class="placeholder">
          {{ t('components.autoprobe.patterns.selectItem') }}
        </div>
      </div>

      <cl-auto-probe-results-container
        v-if="detailNavItem"
        :data="resultsData"
        :fields="resultsFields"
        :active-field-name="resultsActiveField?.name"
        :url="form.url"
        :active-id="activeId"
        @size-change="onSizeChange"
      />
    </div>
  </div>
</template>

<style scoped>
.autoprobe-detail-tab-patterns {
  height: 100%;
  display: flex;

  .content {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;

    .detail-container {
      flex: 1;
      overflow: auto;
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
