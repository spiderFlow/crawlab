import { useStore } from 'vuex';
import { useList } from '@/layouts';

const useAutoProbeList = () => {
  const ns: ListStoreNamespace = 'extract';
  const store = useStore();

  return {
    ...useList<ExtractPattern>(ns),
  };
};

export default useAutoProbeList;
