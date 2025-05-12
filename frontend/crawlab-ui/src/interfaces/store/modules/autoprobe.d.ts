type AutoProbeStoreModule = BaseModule<
  AutoProbeStoreState,
  AutoProbeStoreGetters,
  AutoProbeStoreMutations,
  AutoProbeStoreActions
>;

interface AutoProbeStoreState extends BaseStoreState<AutoProbe> {}

type AutoProbeStoreGetters = BaseStoreGetters<AutoProbe>;

interface AutoProbeStoreMutations extends BaseStoreMutations<AutoProbe> {}

interface AutoProbeStoreActions extends BaseStoreActions<AutoProbe> {}
