import {
  getDefaultStoreActions,
  getDefaultStoreGetters,
  getDefaultStoreMutations,
  getDefaultStoreState,
} from '@/utils/store';
import { TAB_NAME_OVERVIEW, TAB_NAME_TASKS } from '@/constants/tab';
import { translate } from '@/utils/i18n';
import useRequest from '@/services/request';

// i18n
const t = translate;

const { post } = useRequest();

const state = {
  ...getDefaultStoreState<AutoProbe>('autoprobe'),
  tabs: [
    { id: TAB_NAME_OVERVIEW, title: t('common.tabs.overview') },
    { id: TAB_NAME_TASKS, title: t('common.tabs.tasks') },
  ],
} as AutoProbeStoreState;

const getters = {
  ...getDefaultStoreGetters<AutoProbe>(),
} as AutoProbeStoreGetters;

const mutations = {
  ...getDefaultStoreMutations<AutoProbe>(),
} as AutoProbeStoreMutations;

const actions = {
  ...getDefaultStoreActions<AutoProbe>('/ai/autoprobes'),
  runTask: async (
    _: StoreActionContext<AutoProbeStoreState>,
    { id }: { id: string }
  ) => {
    await post(`/ai/autoprobes/${id}/tasks`);
  },
  cancelTask: async (
    _: StoreActionContext<AutoProbeStoreState>,
    { id }: { id: string }
  ) => {
    await post(`/ai/autoprobes/tasks/${id}/cancel`);
  },
} as AutoProbeStoreActions;

export default {
  namespaced: true,
  state,
  getters,
  mutations,
  actions,
} as AutoProbeStoreModule;
