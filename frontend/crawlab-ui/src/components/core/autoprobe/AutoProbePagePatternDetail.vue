<script setup lang="ts">
import { computed } from 'vue';
import { useStore } from 'vuex';
import { translate } from '@/utils';

// i18n
const t = translate;

// props
const props = defineProps<{
  pagePattern: AutoProbeNavItem;
  pageData?: any;
}>();

// store
const store = useStore();
const { autoprobe: state } = store.state as RootStoreState;
const form = computed<AutoProbe>(() => state.form);

// Page pattern data
const pagePatternData = computed(() => form.value?.page_pattern);

// Stats
const fieldCount = computed(() => pagePatternData.value?.fields?.length || 0);
const listCount = computed(() => pagePatternData.value?.lists?.length || 0);
const hasPagination = computed(() => !!pagePatternData.value?.pagination);

// Check if we have top-level properties to display as tables
const topLevelArrays = computed(() => {
  if (!props.pageData || typeof props.pageData !== 'object') return [];
  
  const result = [];
  for (const key in props.pageData) {
    if (Array.isArray(props.pageData[key]) && props.pageData[key].length > 0) {
      // Only include arrays that have objects inside (suitable for tables)
      if (typeof props.pageData[key][0] === 'object' && props.pageData[key][0] !== null) {
        result.push({
          key,
          data: props.pageData[key],
          columns: getColumnsFromData(props.pageData[key])
        });
      }
    }
  }
  
  return result;
});

// Helper function to extract columns from data
function getColumnsFromData(data: any[]): {prop: string; label: string}[] {
  if (!data || !data.length) return [];
  const firstItem = data[0];
  if (typeof firstItem !== 'object' || firstItem === null) return [];
  
  return Object.keys(firstItem).map(key => ({
    prop: key,
    label: key.charAt(0).toUpperCase() + key.slice(1) // Capitalize first letter
  }));
}

// Format page data for display (for properties that aren't array tables)
const formattedPageData = computed(() => {
  if (!props.pageData) return [];
  
  // Convert the page data to a format suitable for display
  const result = [];
  const arrayKeys = topLevelArrays.value.map(item => item.key);
  
  // If it's an object, display key-value pairs (excluding array tables)
  if (typeof props.pageData === 'object' && !Array.isArray(props.pageData)) {
    for (const key in props.pageData) {
      // Skip keys that are already displayed as tables
      if (arrayKeys.includes(key)) continue;
      
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
  // If it's an array but not suitable for tables
  else if (Array.isArray(props.pageData)) {
    try {
      result.push({
        key: 'array',
        value: JSON.stringify(props.pageData, null, 2)
      });
    } catch (e) {
      result.push({
        key: 'array',
        value: String(props.pageData)
      });
    }
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

defineOptions({ name: 'ClAutoProbePagePatternDetail' });
</script>

<template>
  <div class="cl-autoprobe-page-pattern-detail">
    <div class="header">
      <h3>{{ t('components.autoprobe.pagePattern.title') }}</h3>
    </div>
    
    <div v-if="pagePatternData" class="content">
      <el-descriptions :column="1" border>
        <el-descriptions-item :label="t('components.autoprobe.pagePattern.name')">
          {{ pagePatternData.name }}
        </el-descriptions-item>
      </el-descriptions>
      
      <div class="stats-section">
        <h4>{{ t('components.autoprobe.pagePattern.stats') }}</h4>
        <el-row :gutter="20">
          <el-col :span="8">
            <div class="stat-card">
              <div class="stat-value">{{ fieldCount }}</div>
              <div class="stat-label">{{ t('components.autoprobe.pagePattern.fields') }}</div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="stat-card">
              <div class="stat-value">{{ listCount }}</div>
              <div class="stat-label">{{ t('components.autoprobe.pagePattern.lists') }}</div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="stat-card">
              <div class="stat-value">{{ hasPagination ? '✓' : '✗' }}</div>
              <div class="stat-label">{{ t('components.autoprobe.pagePattern.hasPagination') }}</div>
            </div>
          </el-col>
        </el-row>
      </div>
      
      <!-- Array Tables Section - For data that can be displayed as tables -->
      <template v-if="topLevelArrays.length">
        <div v-for="(arrayInfo, index) in topLevelArrays" :key="index" class="page-data-section">
          <h4>{{ arrayInfo.key }}</h4>
          <div class="table">
            <el-table
              :data="arrayInfo.data"
              border
              style="width: 100%"
              height="400"
              highlight-current-row
            >
              <el-table-column
                v-for="column in arrayInfo.columns"
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
                {{ `Total: ${arrayInfo.data.length} ${arrayInfo.data.length === 1 ? 'item' : 'items'}` }}
              </span>
            </div>
          </div>
        </div>
      </template>
      
      <!-- Other Page Data Section - For non-tabular data -->
      <div v-if="formattedPageData.length" class="page-data-section">
        <h4>{{ t('components.autoprobe.pagePattern.otherProperties') || 'Other Properties' }}</h4>
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
      {{ t('components.autoprobe.pagePattern.notFound') || 'Page pattern details not found' }}
    </div>
  </div>
</template>

<style scoped>
.cl-autoprobe-page-pattern-detail {
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