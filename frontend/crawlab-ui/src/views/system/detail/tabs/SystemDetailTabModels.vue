<script setup lang="tsx">
import { ref, computed, onBeforeMount } from 'vue';
import { ElSpace, ElMessage, ElMessageBox } from 'element-plus';
import { ClSwitch, ClTag, ClNavLink, ClIcon } from '@/components';
import useRequest from '@/services/request';
import { getDefaultPagination, plainClone, translate } from '@/utils';
import {
  ACTION_DELETE,
  ACTION_EDIT,
  ACTION_VIEW,
  TABLE_COLUMN_NAME_ACTIONS,
} from '@/constants';
import { getLLMProviderItems } from '@/utils/ai';

const t = translate;

const { getList, put, post, del } = useRequest();

const llmProviders = ref<LLMProvider[]>([]);
const llmProvidersTotal = ref(0);
const form = ref<LLMProvider>();
const formRef = ref();

const getLlmProviderItem = (type: LLMProviderType) => {
  return getLLMProviderItems().find(item => item.type === type);
};

const getLLMProviderList = async () => {
  const res = await getList('/ai/llm/providers');
  llmProviders.value = res.data || [];
  llmProvidersTotal.value = res.total || 0;
};
const onClickAdd = () => {
  form.value = {
    name: '',
    enabled: true,
    models: [],
  };
  dialogVisible.value = true;
};
const onClose = () => {
  dialogVisible.value = false;
  form.value = undefined;
};
const onConfirm = async () => {
  if (!form.value) return;
  if (!formRef.value) return;
  await formRef.value.validate();
  const { _id } = form.value;
  try {
    dialogConfirmLoading.value = true;
    if (_id) {
      await put(`/ai/llm/providers/${_id}`, {
        data: form.value,
      });
    } else {
      await post('/ai/llm/providers', {
        data: form.value,
      });
    }
    dialogVisible.value = false;
    form.value = undefined;
    ElMessage.success(t('common.message.success.save'));
    await getLLMProviderList();
  } finally {
    dialogConfirmLoading.value = false;
  }
};
const deleteByIdConfirm = async (id: string) => {
  await ElMessageBox.confirm(
    t('common.messageBox.confirm.delete'),
    t('common.actions.delete'),
    {
      type: 'warning',
      confirmButtonClass: 'el-button--danger delete-confirm-btn',
    }
  );
  try {
    await del(`/ai/llm/providers/${id}`);
    ElMessage.success(t('common.message.success.delete'));
    await getLLMProviderList();
  } catch (e) {
    ElMessage.error((e as Error).message);
  }
};

const tableColumns = computed<TableColumns<LLMProvider>>(() => {
  return [
    {
      key: 'name',
      label: t('views.system.ai.name'),
      width: '200px',
      value: (row: LLMProvider) => {
        const item = getLlmProviderItem(row.type!);
        return (
          <ClNavLink
            onClick={() => {
              form.value = row;
              dialogVisible.value = true;
            }}
          >
            <span style="margin-right: 5px">
              <ClIcon icon={item?.icon} />
            </span>
            <span>{row.name}</span>
          </ClNavLink>
        );
      },
    },
    {
      key: 'enabled',
      label: t('views.system.ai.enabled'),
      width: '90px',
      value: (row: LLMProvider) => {
        return (
          <ClSwitch
            modelValue={row.enabled}
            onChange={async (enabled: boolean) => {
              const originalEnabled = row.enabled;
              row.enabled = enabled;
              try {
                await put(`/ai/llm/providers/${row._id}`, {
                  data: { ...row, enabled },
                });
                ElMessage.success(
                  t(
                    `common.message.success.${enabled ? 'enabled' : 'disabled'}`
                  )
                );
              } catch (e) {
                ElMessage.error((e as Error).message);
                row.enabled = originalEnabled;
              }
            }}
          />
        );
      },
    },
    {
      key: 'models',
      label: t('views.system.ai.models'),
      width: 'auto',
      minWidth: '300px',
      value: (row: LLMProvider) => {
        return (
          <ElSpace direction="horizontal" gap={8} wrap>
            {row.models?.map(model => {
              return <ClTag label={model} />;
            })}
          </ElSpace>
        );
      },
    },
    {
      key: TABLE_COLUMN_NAME_ACTIONS,
      label: t('components.table.columns.actions'),
      fixed: 'right',
      width: '150',
      buttons: [
        {
          tooltip: t('common.actions.edit'),
          onClick: row => {
            form.value = plainClone(row);
            dialogVisible.value = true;
          },
          action: ACTION_EDIT,
        },
        {
          tooltip: t('common.actions.delete'),
          disabled: row => row.spiders > 0,
          onClick: row => deleteByIdConfirm(row._id),
          action: ACTION_DELETE,
          contextMenu: true,
        },
      ],
      disableTransfer: true,
    },
  ] as TableColumns<LLMProvider>;
});
const tablePagination = ref<TablePagination>(getDefaultPagination());
const tableData = computed(() => llmProviders.value);
const tableTotal = computed(() => llmProvidersTotal.value);
const onTablePaginationChange = async (pagination: TablePagination) => {
  tablePagination.value = pagination;
  await getLLMProviderList();
};

const dialogVisible = ref(false);
const dialogTitle = computed(() => {
  if (form.value?._id) {
    return t('views.system.ai.actions.edit.llmProvider');
  }
  return t('views.system.ai.actions.new.llmProvider');
});
const dialogConfirmLoading = ref(false);

onBeforeMount(getLLMProviderList);

defineOptions({ name: 'ClSystemDetailTabModels' });
</script>

<template>
  <div class="models-container">
    <cl-nav-actions>
      <cl-nav-action-group>
        <cl-nav-action-item>
          <cl-nav-action-button
            button-type="label"
            :icon="['fa', 'plus']"
            :label="t('views.system.ai.actions.new.llmProvider')"
            @click="onClickAdd"
          />
        </cl-nav-action-item>
      </cl-nav-action-group>
    </cl-nav-actions>
    <cl-table
      :columns="tableColumns"
      :data="tableData"
      :page="tablePagination.page"
      :page-size="tablePagination.size"
      :total="tableTotal"
      embedded
      @pagination-change="onTablePaginationChange"
    >
      <template #empty>
        <cl-label-button
          :icon="['fa', 'plus']"
          :label="t('views.system.ai.actions.new.llmProvider')"
          @click="onClickAdd"
        />
      </template>
    </cl-table>

    <cl-dialog
      ref="dialogRef"
      :visible="dialogVisible"
      :title="dialogTitle"
      :confirm-loading="dialogConfirmLoading"
      @confirm="onConfirm"
      @close="onClose"
    >
      <cl-llm-provider-form
        ref="formRef"
        v-model="form"
        :loading="dialogConfirmLoading"
      />
    </cl-dialog>
  </div>
</template>

<style scoped>
.models-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}
</style>
