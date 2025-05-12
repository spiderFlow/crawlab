import { Store } from 'vuex';
import { getDefaultService } from '@/utils';

const useAutoProbeService = (
  store: Store<RootStoreState>
): Services<AutoProbe> => {
  const ns: ListStoreNamespace = 'autoprobe';

  return {
    ...getDefaultService<AutoProbe>(ns, store),
  };
};

export default useAutoProbeService;
