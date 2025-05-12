import { TAB_NAME_OVERVIEW, TAB_NAME_TASKS } from '@/constants/tab';
import {
  ClAutoProbeList,
  ClAutoProbeDetail,
  ClAutoProbeDetailTabOverview,
  ClAutoProbeDetailTabTasks,
} from '@/views';
import { getIconByTabName, translate } from '@/utils';
import { RouteLocation } from 'vue-router';

const t = translate;

const endpoint = '/autoprobes';

export default [
  {
    routeConcept: 'autoprobe',
    name: 'AutoProbeList',
    path: endpoint,
    title: t('layouts.routes.autoprobe.list.title'),
    component: async () => ClAutoProbeList,
  },
  {
    routeConcept: 'autoprobe',
    name: 'AutoProbeDetail',
    path: `${endpoint}/:id`,
    title: t('layouts.routes.autoprobe.detail.title'),
    redirect: (to: RouteLocation) => {
      return { path: to.path + '/overview' };
    },
    component: async () => ClAutoProbeDetail,
    children: [
      {
        path: TAB_NAME_OVERVIEW,
        title: t('layouts.routes.autoprobe.detail.tabs.overview'),
        icon: getIconByTabName(TAB_NAME_OVERVIEW),
        component: async () => ClAutoProbeDetailTabOverview,
      },
      {
        path: TAB_NAME_TASKS,
        title: t('layouts.routes.autoprobe.detail.tabs.tasks'),
        icon: getIconByTabName(TAB_NAME_TASKS),
        component: async () => ClAutoProbeDetailTabTasks,
      },
    ],
  },
] as Array<ExtendedRouterRecord>;
