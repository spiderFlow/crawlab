<script setup lang="ts">
import { computed } from 'vue';
import { useStore } from 'vuex';
import { translate } from '@/utils';

// i18n
const t = translate;

// props
const props = defineProps<{
  pagination: AutoProbeNavItem;
  pageData?: any;
}>();

// store
const store = useStore();
const { autoprobe: state } = store.state as RootStoreState;
const form = computed<AutoProbe>(() => state.form);

// Pagination data
const paginationData = computed(() => form.value?.page_pattern?.pagination);

// Format page data for display
const formattedPageData = computed(() => {
  if (!props.pageData) return null;
  
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

defineOptions({ name: 'ClAutoProbePaginationDetail' });
</script>

<template>
  <div class="cl-autoprobe-pagination-detail">
    <div class="header">
      <h3>{{ t('components.autoprobe.pagination.title') }}</h3>
    </div>
    
    <div v-if="paginationData" class="content">
      <el-descriptions :column="1" border>
        <el-descriptions-item :label="t('components.autoprobe.pagination.type')">
          {{ paginationData.type }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.pagination.selectorType')" v-if="paginationData.selector_type">
          {{ paginationData.selector_type }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.pagination.selector')" v-if="paginationData.selector">
          {{ paginationData.selector }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.pagination.maxPages')" v-if="paginationData.max_pages">
          {{ paginationData.max_pages }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('components.autoprobe.pagination.startPage')" v-if="paginationData.start_page">
          {{ paginationData.start_page }}
        </el-descriptions-item>
      </el-descriptions>
      
      <!-- Page Data Section -->
      <div v-if="pageData" class="page-data-section">
        <h4>{{ t('components.autoprobe.pagination.pageData') || 'Page Data' }}</h4>
        <el-card shadow="never" class="page-data-card">
          <pre v-if="typeof formattedPageData === 'string' && formattedPageData.startsWith('{')" class="json-value">{{ formattedPageData }}</pre>
          <pre v-else-if="typeof formattedPageData === 'string' && formattedPageData.startsWith('[')" class="json-value">{{ formattedPageData }}</pre>
          <span v-else>{{ formattedPageData }}</span>
        </el-card>
      </div>
    </div>
    
    <div v-else class="not-found">
      {{ t('components.autoprobe.pagination.notFound') || 'Pagination details not found' }}
    </div>
  </div>
</template>

<style scoped>
.cl-autoprobe-pagination-detail {
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
    
    .json-value {
      margin: 0;
      white-space: pre-wrap;
      word-break: break-word;
      font-family: monospace;
      font-size: 12px;
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