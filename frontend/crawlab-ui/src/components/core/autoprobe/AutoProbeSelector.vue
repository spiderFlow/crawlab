<script setup lang="ts">
import { computed } from 'vue';
import { translate } from '@/utils';

const props = defineProps<{
  type: 'field' | 'pagination';
  rule: FieldRule | Pagination;
}>();

const t = translate;

const selectorIcon = computed<Icon>(() => {
  const { rule } = props;
  switch (rule.selector_type) {
    case 'css':
      return ['fab', 'css'];
    case 'xpath':
      return ['fa', 'code'];
    case 'regex':
      return ['fa', 'search'];
  }
});

const selectorType = computed(() => {});

const extractionIcon = computed<Icon>(() => {
  const { rule, type } = props;
  if (type === 'field') {
    switch ((rule as FieldRule).extraction_type) {
      case 'attribute':
        return ['fa', 'tag'];
      case 'text':
        return ['fa', 'font'];
      case 'html':
        return ['fa', 'code'];
      default:
        return ['fa', 'question'];
    }
  } else {
    return ['fa', 'question'];
  }
});

const extractionLabel = computed(() => {
  const { rule, type } = props;
  if (type === 'field') {
    return (rule as FieldRule).extraction_type;
  } else {
    return '';
  }
});

defineOptions({ name: 'ClAutoProbeSelector' });
</script>

<template>
  <div class="selector">
    <cl-tag
      :icon="selectorIcon"
      :label="rule.selector"
      :tooltip="
        t(
          `components.autoprobe.pagePattern.selectorTypes.${rule.selector_type}`
        )
      "
    />
    <template v-if="type === 'field'">
      <span class="divider">
        <cl-icon :icon="['fa', 'angle-right']" />
      </span>
      <cl-tag :icon="extractionIcon" :label="extractionLabel" />
    </template>
  </div>
</template>

<style scoped>
.selector {
  display: flex;
  align-items: center;
  gap: 5px;

  .divider {
    color: var(--el-text-color-secondary);
    font-size: 10px;
  }
}
</style>
