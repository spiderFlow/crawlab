import { useStore } from 'vuex';
import { useList } from '@/layouts';

const useAutoProbeList = () => {
  const ns: ListStoreNamespace = 'autoprobe';
  const store = useStore();

  return {
    ...useList<AutoProbe>(ns, store),
  };
};

export default useAutoProbeList;
