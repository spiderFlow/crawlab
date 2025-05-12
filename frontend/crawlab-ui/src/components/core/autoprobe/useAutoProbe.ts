import { computed } from 'vue';
import { Store } from 'vuex';
import { useForm } from '@/components';
import useAutoProbeService from '@/services/autoprobe/autoprobeService';
import { getDefaultFormComponentData } from '@/utils/form';

// form component data
const formComponentData = getDefaultFormComponentData<AutoProbe>();

const useAutoProbe = (store: Store<RootStoreState>) => {
  // store
  const ns = 'autoprobe';
  const state = store.state[ns];

  // form rules
  const formRules: FormRules = {};

  // all autoprobe select options
  const allAutoProbeSelectOptions = computed<SelectOption[]>(() =>
    state.allList.map(d => {
      return {
        label: d.name,
        value: d._id,
      };
    })
  );

  return {
    ...useForm<AutoProbe>(
      'autoprobe',
      store,
      useAutoProbeService(store),
      formComponentData
    ),
    formRules,
    allAutoProbeSelectOptions,
  };
};

export default useAutoProbe;
