<script setup lang="ts">
import { computed } from 'vue';
import { useStore } from 'vuex';
import { translate } from '@/utils';

// i18n
const t = translate;

// props
const props = defineProps<{
  list: AutoProbeNavItem;
  pageData?: any;
}>();

// store
const store = useStore();
const { autoprobe: state } = store.state as RootStoreState;
const form = computed<AutoProbe>(() => state.form);

// Find the actual list data from the form based on the nav item
const listData = computed<ListRule | undefined>(() => {
  if (!form.value?.page_pattern) return undefined;
  
  const listName = props.list.id;
  
  // Find the list recursively
  const findList = (lists: ListRule[] | undefined): ListRule | undefined => {
    if (!lists) return undefined;
    for (const list of lists) {
      if (list.name === listName) return list;
      const nestedResult = findList(list.item_pattern?.lists);
      if (nestedResult) return nestedResult;
    }
    return undefined;
  };
  
  return findList(form.value.page_pattern.lists);
});

// Get raw array data for direct table display (when page data is an array)
const tableData = computed(() => {
  if (!props.pageData) return [];
  
  // If page data is an array, it's likely the list data itself
  if (Array.isArray(props.pageData)) {
    return props.pageData;
  }
  // If it's an object with a property matching the list id, try that
  else if (typeof props.pageData === 'object' && props.pageData !== null && 
           Array.isArray(props.pageData[props.list.id])) {
    return props.pageData[props.list.id];
  }
  
  return [];
});

// Get table columns dynamically based on data
const tableColumns = computed(() => {
  if (!tableData.value.length) return [];
  
  // Get the first item to determine columns
  const firstItem = tableData.value[0];
  if (typeof firstItem !== 'object' || firstItem === null) {
    // Simple value array
    return [{ prop: 'value', label: 'Value' }];
  }
  
  // Extract columns from object properties
  return Object.keys(firstItem).map(key => ({
    prop: key,
    label: key.charAt(0).toUpperCase() + key.slice(1) // Capitalize first letter
  }));
});

// Determine if we should show the data table
const showDataTable = computed(() => {
  return tableData.value.length > 0 && tableColumns.value.length > 0;
});

// Format page data for display (key-value format for non-array data)
const formattedPageData = computed(() => {
  if (!props.pageData || showDataTable.value) return [];
  
  // Convert the page data to a format suitable for display
  const result = [];
  
  // If it's an object, display key-value pairs
  if (typeof props.pageData === 'object' && !Array.isArray(props.pageData)) {
    for (const key in props.pageData) {
      if (Object.prototype.hasOwnProperty.call(props.pageData, key)) {
        const value = props.pageData[key];
        let displayValue = value;
        
        // Format based on value type
        if (typeof value === 'object') {
          try {
            displayValue = JSON.stringify(value, null, 2);
          } catch (e) {
            displayValue = String(value);
          }
        }
        
        result.push({
          key,
          value: displayValue
        });
      }
    }
  } 
  // If it's an array but not suitable for table display, process each item
  else if (Array.isArray(props.pageData)) {
    props.pageData.forEach((item, index) => {
      let displayValue = item;
      
      // Format based on type
      if (typeof item === 'object') {
        try {
          displayValue = JSON.stringify(item, null, 2);
        } catch (e) {
          displayValue = String(item);
        }
      }
      
      result.push({
        key: `Item ${index + 1}`,
        value: displayValue
      });
    });
  }
  // For primitive types
  else {
    result.push({
      key: 'value',
      value: props.pageData
    });
  }
  
  return result;
});

// Compute fields and nested lists for display
const fields = computed(() => listData.value?.item_pattern?.fields || []);
const nestedLists = computed(() => listData.value?.item_pattern?.lists || []);

// Display value for list selector - if empty, it points to itself (self)
const displayListSelector = computed(() => {
  if (!listData.value?.list_selector) return t('components.autoprobe.list.self') || 'self';
  return listData.value.list_selector;
});

// Display value for item selector - if empty, it points to itself (self)
const displayItemSelector = computed(() => {
  if (!listData.value?.item_selector) return t('components.autoprobe.list.self') || 'self';
  return listData.value.item_selector;
});

// Process fields for display, marking empty selectors as "self"
const processedFields = computed(() => {
  return fields.value.map(field => ({
    ...field,
    displaySelector: field.selector || (t('components.autoprobe.field.self') || 'self')
  }));
});

// Process nested lists for display, marking empty selectors as "self"
const processedNestedLists = computed(() => {
  return nestedLists.value.map(nestedList => ({
    ...nestedList,
    displayListSelector: nestedList.list_selector || (t('components.autoprobe.list.self') || 'self')
  }));
});

defineOptions({ name: 'ClAutoProbeListDetail' });
</script>

<template>
  <div class="cl-autoprobe-list-detail">
    <div class="header">
      <h3>{{ t('components.autoprobe.list.title') }}: {{ list.label }}</h3>
    </div>
    
    <div v-if="listData" class="content">
      <el-descriptions :column="1" border>
        <el-descriptions-item :label="t('components.autoprobe.list.name')">
          {{ listData.name }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.list.listSelector')">
          {{ displayListSelector }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.list.listSelectorType')">
          {{ listData.list_selector_type }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.list.itemSelector')">
          {{ displayItemSelector }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.list.itemSelectorType')">
          {{ listData.item_selector_type }}
        </el-descriptions-item>
      </el-descriptions>
      
      <div v-if="fields.length" class="fields-section">
        <h4>{{ t('components.autoprobe.list.fields') }}</h4>
        <el-table :data="processedFields" border style="width: 100%">
          <el-table-column prop="name" :label="t('components.autoprobe.field.name')" width="200" />
          <el-table-column :label="t('components.autoprobe.field.selector')">
            <template #default="scope">
              {{ scope.row.displaySelector }}
            </template>
          </el-table-column>
        </el-table>
      </div>
      
      <div v-if="nestedLists.length" class="nested-lists-section">
        <h4>{{ t('components.autoprobe.list.nestedLists') }}</h4>
        <el-table :data="processedNestedLists" border style="width: 100%">
          <el-table-column prop="name" :label="t('components.autoprobe.list.name')" width="200" />
          <el-table-column :label="t('components.autoprobe.list.listSelector')">
            <template #default="scope">
              {{ scope.row.displayListSelector }}
            </template>
          </el-table-column>
        </el-table>
      </div>
      
      <!-- Dynamic Data Table for List Items -->
      <div v-if="showDataTable" class="page-data-section">
        <h4>{{ t('components.autoprobe.list.items') || 'List Items' }}</h4>
        <div class="table">
          <el-table
            :data="tableData"
            border
            style="width: 100%"
            height="400"
            highlight-current-row
          >
            <el-table-column
              v-for="column in tableColumns"
              :key="column.prop"
              :prop="column.prop"
              :label="column.label"
              sortable
            >
              <template #default="scope">
                <template v-if="typeof scope.row[column.prop] !== 'object' || scope.row[column.prop] === null">
                  {{ scope.row[column.prop] }}
                </template>
                <pre v-else class="json-value">{{ JSON.stringify(scope.row[column.prop], null, 2) }}</pre>
              </template>
            </el-table-column>
          </el-table>
          <div class="table-footer">
            <span class="total-items">
              {{ `Total: ${tableData.length} ${tableData.length === 1 ? 'item' : 'items'}` }}
            </span>
          </div>
        </div>
      </div>
      
      <!-- Page Data Section (for non-tabular data) -->
      <div v-else-if="formattedPageData.length" class="page-data-section">
        <h4>{{ t('components.autoprobe.list.pageData') || 'Page Data' }}</h4>
        <el-table :data="formattedPageData" border style="width: 100%">
          <el-table-column prop="key" :label="t('components.autoprobe.pageData.key') || 'Key'" width="200" />
          <el-table-column :label="t('components.autoprobe.pageData.value') || 'Value'">
            <template #default="scope">
              <pre v-if="typeof scope.row.value === 'string' && (scope.row.value.startsWith('{') || scope.row.value.startsWith('['))" class="json-value">{{ scope.row.value }}</pre>
              <span v-else>{{ scope.row.value }}</span>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
    
    <div v-else class="not-found">
      {{ t('components.autoprobe.list.notFound') || 'List details not found' }}
    </div>
  </div>
</template>

<style scoped>
.cl-autoprobe-list-detail {
  padding: 16px;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  
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
    
    .fields-section,
    .nested-lists-section,
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