<script setup lang="ts">
import { computed } from 'vue';
import { useStore } from 'vuex';
import { translate } from '@/utils';

// i18n
const t = translate;

// props
const props = defineProps<{
  field: AutoProbeNavItem;
  pageData?: any;
}>();

// store
const store = useStore();
const { autoprobe: state } = store.state as RootStoreState;
const form = computed<AutoProbe>(() => state.form);

// Find the actual field data from the form based on the nav item
const fieldData = computed<FieldRule | undefined>(() => {
  if (!form.value?.page_pattern) return undefined;
  
  // If it's a direct field of the page pattern
  if (props.field.id.indexOf('-') === -1) {
    return form.value.page_pattern.fields?.find(f => f.name === props.field.label);
  }
  
  // If it's a field within a list, parse the ID
  const [listName, fieldName] = props.field.id.split('-');
  
  // Find the list first
  const findList = (lists: ListRule[] | undefined): ListRule | undefined => {
    if (!lists) return undefined;
    for (const list of lists) {
      if (list.name === listName) return list;
      const nestedResult = findList(list.item_pattern?.lists);
      if (nestedResult) return nestedResult;
    }
    return undefined;
  };
  
  const list = findList(form.value.page_pattern.lists);
  if (!list || !list.item_pattern?.fields) return undefined;
  
  return list.item_pattern.fields.find(f => f.name === fieldName);
});

// Display value for selector - if empty, it points to itself (self)
const displaySelector = computed(() => {
  if (!fieldData.value?.selector) return t('components.autoprobe.field.self') || 'self';
  return fieldData.value.selector;
});

// Determine if we have tabular data (array of objects)
const isTabularData = computed(() => {
  if (!props.pageData) return false;
  
  // Check if it's an array with objects inside
  if (Array.isArray(props.pageData) && props.pageData.length > 0 && 
      typeof props.pageData[0] === 'object' && props.pageData[0] !== null) {
    return true;
  }
  
  return false;
});

// Get table data for tabular display
const tableData = computed(() => {
  if (!props.pageData || !isTabularData.value) return [];
  return props.pageData;
});

// Get table columns
const tableColumns = computed(() => {
  if (!tableData.value.length) return [];
  
  // Get the first item to extract columns
  const firstItem = tableData.value[0];
  if (typeof firstItem !== 'object' || firstItem === null) return [];
  
  // Extract columns from object keys
  return Object.keys(firstItem).map(key => ({
    prop: key,
    label: key.charAt(0).toUpperCase() + key.slice(1) // Capitalize first letter
  }));
});

// Format page data for display (for non-tabular data)
const formattedPageData = computed(() => {
  if (!props.pageData || isTabularData.value) return null;
  
  let displayValue = props.pageData;
  
  // Format based on type
  if (typeof displayValue === 'object') {
    // For object or array, stringify with formatting
    try {
      displayValue = JSON.stringify(displayValue, null, 2);
    } catch (e) {
      displayValue = String(displayValue);
    }
  }
  
  return displayValue;
});

defineOptions({ name: 'ClAutoProbeFieldDetail' });
</script>

<template>
  <div class="cl-autoprobe-field-detail">
    <div class="header">
      <h3>{{ t('components.autoprobe.field.title') }}: {{ field.label }}</h3>
    </div>
    
    <div v-if="fieldData" class="content">
      <el-descriptions :column="1" border>
        <el-descriptions-item :label="t('components.autoprobe.field.name')">
          {{ fieldData.name }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.field.selector')">
          {{ displaySelector }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.field.type')">
          {{ fieldData.selector_type }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.field.extractionType')">
          {{ fieldData.extraction_type }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.field.attributeName')" v-if="fieldData.attribute_name">
          {{ fieldData.attribute_name }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.field.defaultValue')" v-if="fieldData.default_value">
          {{ fieldData.default_value }}
        </el-descriptions-item>
      </el-descriptions>
      
      <!-- Tabular Data Display -->
      <div v-if="isTabularData" class="page-data-section">
        <h4>{{ t('components.autoprobe.field.data') || 'Field Data' }}</h4>
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
      
      <!-- Single Value or Non-tabular Data -->
      <div v-else-if="pageData" class="page-data-section">
        <h4>{{ t('components.autoprobe.field.pageData') || 'Page Data' }}</h4>
        <el-card shadow="never" class="page-data-card">
          <pre v-if="typeof formattedPageData === 'string' && (formattedPageData.startsWith('{') || formattedPageData.startsWith('['))" class="json-value">{{ formattedPageData }}</pre>
          <span v-else>{{ formattedPageData }}</span>
        </el-card>
      </div>
    </div>
    
    <div v-else class="not-found">
      {{ t('components.autoprobe.field.notFound') || 'Field details not found' }}
    </div>
  </div>
</template>

<style scoped>
.cl-autoprobe-field-detail {
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
    
    .page-data-section {
      margin-top: 20px;
      
      h4 {
        margin: 0 0 12px 0;
        font-size: 14px;
        font-weight: 500;
      }
      
      .page-data-card {
        border: 1px solid var(--el-border-color-light);
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