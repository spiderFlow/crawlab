export declare global {
  interface AutoProbe extends BaseModel {
    name?: string;
    url?: string;
    query?: string;
    last_task_id?: string;
    last_task?: AutoProbeTask;
    default_task_id?: string;
    page_pattern?: PagePattern;
    page_data?: PageDataRow;
  }

  type AutoProbeTaskStatus =
    | 'pending'
    | 'running'
    | 'completed'
    | 'failed'
    | 'cancelled';

  type SelectorType = 'css' | 'xpath' | 'regex';
  type ExtractType = 'text' | 'attribute' | 'html';

  interface BaseSelector {
    name: string;
    selector_type: SelectorType;
    selector: string;
  }

  interface FieldRule extends BaseSelector {
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

  type PaginationRule = BaseSelector;

  interface PagePattern {
    name: string;
    fields?: FieldRule[];
    lists?: ListRule[];
    pagination?: PaginationRule;
  }

  type PageData = Record<string, string | number | boolean | PageData[]>;

  interface AutoProbeTask extends BaseModel {
    autoprobe_id: string;
    url?: string;
    query?: string;
    status: AutoProbeTaskStatus;
    error?: string;
    page_pattern?: PagePattern;
    page_data?: PageData;
    provider_id?: string;
    model?: string;
    usage?: LLMResponseUsage;
  }

  type AutoProbeItemType = 'page_pattern' | 'list' | 'field' | 'pagination';

  interface AutoProbeNavItem<T = any> extends NavItem<T> {
    name?: string;
    type?: AutoProbeItemType;
    rule?: ListRule | FieldRule | PaginationRule;
    children?: AutoProbeNavItem[];
    parent?: AutoProbeNavItem;
    fieldCount?: number;
  }
}
