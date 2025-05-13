<script setup lang="ts">
const props = defineProps<{
  fieldRule: FieldRule;
}>();

defineOptions({ name: 'ClAutoProbeFieldRule' });
</script>

<template>
  <div class="field-rule">
    <div class="name">
      <span class="icon">
        <cl-icon :icon="['fa', 'share-nodes']" />
      </span>
      <label>{{ fieldRule.name }}</label>
    </div>
    <div class="selector">
      <span>{{ fieldRule.selector_type }}</span>
      <span>{{ fieldRule.selector }}</span>
    </div>
    <div class="extraction">
      <span>{{ fieldRule.extraction_type }}</span>
      <span>{{ fieldRule.attribute_name }}</span>
    </div>
  </div>
</template>

<style scoped>
.field-rule {
  display: flex;
  align-items: center;

  .name {
    flex: 0 0 180px;
    display: flex;
    align-items: center;

    .icon {
      margin-right: 5px;
    }
  }

  .selector {
    display: flex;
    align-items: center;
  }

  .extraction {
    display: flex;
    align-items: center;
  }
}
</style>
