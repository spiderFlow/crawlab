const autoprobe: LComponentsAutoProbe = {
  form: {
    name: 'Name',
    url: 'URL',
    query: 'Query',
  },
  task: {
    status: {
      label: {
        pending: 'Pending',
        running: 'Running',
        completed: 'Completed',
        failed: 'Failed',
        cancelled: 'Cancelled',
        unknown: 'Unknown',
      },
      tooltip: {
        pending: 'The task is waiting to be processed',
        running: 'The task is currently running',
        completed: 'The task has been completed successfully',
        failed: 'The task has failed',
        cancelled: 'The task has been cancelled',
        unknown: 'The task status is unknown',
      },
    },
  },
};

export default autoprobe;
