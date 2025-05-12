interface LComponentsAutoProbe {
  form: {
    name: string;
    url: string;
    query: string;
  };
  task: {
    status: {
      label: {
        pending: string;
        running: string;
        completed: string;
        failed: string;
        cancelled: string;
        unknown: string;
      };
      tooltip: {
        pending: string;
        running: string;
        completed: string;
        failed: string;
        cancelled: string;
        unknown: string;
      };
    };
  };
}
