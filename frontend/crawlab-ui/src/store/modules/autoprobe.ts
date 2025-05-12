import {
  getDefaultStoreActions,
  getDefaultStoreGetters,
  getDefaultStoreMutations,
  getDefaultStoreState,
} from '@/utils/store';
import { TAB_NAME_OVERVIEW } from '@/constants/tab';
import { translate } from '@/utils/i18n';

// i18n
const t = translate;

const state = {
  ...getDefaultStoreState<AutoProbe>('autoprobe'),
  tabs: [
    { id: TAB_NAME_OVERVIEW, title: t('common.tabs.overview') },
  ],
} as AutoProbeStoreState;

const getters = {
  ...getDefaultStoreGetters<AutoProbe>(),
} as AutoProbeStoreGetters;

const mutations = {
  ...getDefaultStoreMutations<AutoProbe>(),
} as AutoProbeStoreMutations;

const actions = {
  ...getDefaultStoreActions<AutoProbe>('/projects'),
} as AutoProbeStoreActions;

export default {
  namespaced: true,
  state,
  getters,
  mutations,
  actions,
} as AutoProbeStoreModule;
