<script setup lang="ts">
import { useStore } from 'vuex';
import { useAutoProbe } from '@/components';
import { translate } from '@/utils';

// i18n
const t = translate;

// store
const store = useStore();
const { form, formRef, formRules, isSelectiveForm, isFormItemDisabled } =
  useAutoProbe(store);
defineOptions({ name: 'ClAutoProbeForm' });
</script>

<template>
  <cl-form
    v-if="form"
    ref="formRef"
    :model="form"
    :rules="formRules"
    :selective="isSelectiveForm"
  >
    <cl-form-item
      :span="2"
      :offset="2"
      :label="t('components.project.form.name')"
      not-editable
      prop="name"
      required
    >
      <el-input
        v-model="form.name"
        :disabled="isFormItemDisabled('name')"
        :placeholder="t('components.autoprobe.form.name')"
      />
    </cl-form-item>
    <cl-form-item
      :span="4"
      :label="t('components.autoprobe.form.url')"
      prop="url"
      required
    >
      <el-input
        v-model="form.url"
        :disabled="isFormItemDisabled('url')"
        :placeholder="t('components.autoprobe.form.url')"
        type="textarea"
      />
    </cl-form-item>
    <cl-form-item
      :span="4"
      :label="t('components.autoprobe.form.query')"
      prop="query"
    >
      <el-input
        v-model="form.query"
        :disabled="isFormItemDisabled('query')"
        :placeholder="t('components.autoprobe.form.query')"
        type="textarea"
      />
    </cl-form-item>
  </cl-form>
</template>
