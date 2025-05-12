export declare global {
  interface AutoProbe extends BaseModel {
    name?: string;
    url?: string;
    query?: string;
  }

  type AutoProbeTaskStatus = 'pending' | 'running' | 'completed' | 'failed';

  type SelectorType = 'css' | 'xpath' | 'regex';
  type ExtractType = 'text' | 'attribute' | 'html';
  type PaginationType = 'next' | 'load' | 'scroll';

  interface FieldRule {
    name: string;
    selector_type: SelectorType;
    selector: string;
    extraction_type: ExtractType;
    attribute_name?: string;
    default_value?: string;
  }

  interface ItemPattern {
    fields?: FieldRule[];
    lists?: ListRule[];
  }

  interface ListRule {
    name: string;
    list_selector_type: SelectorType;
    list_selector: string;
    item_selector_type: SelectorType;
    item_selector: string;
    item_pattern: ItemPattern;
  }

  interface Pagination {
    type: PaginationType;
    selector_type?: SelectorType;
    selector?: string;
    max_pages?: number;
    start_page?: number;
  }

  interface PagePattern {
    name: string;
    fields?: FieldRule[];
    lists?: ListRule[];
    pagination?: Pagination;
  }

  interface PageData {
    data?: Record<string, any>;
    list_data?: any[][];
  }

  interface AutoProbeTask extends BaseModel {
    autoprobe_id: string;
    url?: string;
    query?: string;
    status: AutoProbeTaskStatus;
    page_pattern?: PagePattern;
    page_data?: PageData;
    provider_id?: string;
    model?: string;
    usage?: LLMResponseUsage;
  }

  interface AutoProbeFetchResult extends BaseModel {
    autoprobe_id: string;
    url: string;
    html?: string;
  }
}
