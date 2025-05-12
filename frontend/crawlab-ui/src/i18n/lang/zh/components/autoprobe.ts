const autoprobe: LComponentsAutoProbe = {
  form: {
    name: '名称',
    url: 'URL',
    query: '查询',
  },
  task: {
    status: {
      label: {
        pending: '等待中',
        running: '运行中',
        completed: '已完成',
        failed: '失败',
        cancelled: '已取消',
        unknown: '未知',
      },
      tooltip: {
        pending: '任务正在等待处理',
        running: '任务正在运行',
        completed: '任务已成功完成',
        failed: '任务执行失败',
        cancelled: '任务已被取消',
        unknown: '任务状态未知',
      },
    },
  },
};

export default autoprobe;
