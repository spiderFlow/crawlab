interface LViewsAutoProbe {
  table: {
    columns: {
      name: string;
      url: string;
      query: string;
      status: string;
    };
  };
  navActions: LNavActions & {
    run: {
      label: string;
      tooltip: string;
    };
  };
}
