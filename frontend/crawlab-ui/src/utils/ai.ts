export const DEFAULT_OPENAI_MODELS = [
  'gpt-4o',
  'gpt-4o-mini',
  'gpt-4.1',
  'gpt-4.1-mini',
  'gpt-4.1-nano',
  'o1-mini',
  'o1',
  'o3-mini',
  'o3',
];

export const getLLMProviderItems = (): LLMProviderItem[] => {
  return [
    {
      key: 'openai',
      name: 'OpenAI',
      icon: ['svg', 'openai'],
      defaultModels: DEFAULT_OPENAI_MODELS,
    },
    {
      key: 'azure-openai',
      name: 'Azure OpenAI',
      icon: ['svg', 'azure'],
      defaultModels: DEFAULT_OPENAI_MODELS,
      defaultApiVersions: ['2025-02-01-preview'],
    },
    {
      key: 'anthropic',
      name: 'Anthropic',
      icon: ['svg', 'anthropic'],
      defaultModels: [
        'claude-3-7-sonnet-latest',
        'claude-3-5-haiku-latest',
        'claude-3-5-sonnet-latest',
        'claude-3-opus-latest',
      ],
    },
    {
      key: 'gemini',
      name: 'Gemini',
      icon: ['svg', 'gemini'],
      defaultModels: [
        'gemini-2.5-pro-preview-03-25',
        'gemini-2.0-flash',
        'gemini-2.0-flash-lite',
        'gemini-1.5-flash',
        'gemini-1.5-flash-8b',
        'gemini-1.5-pro',
      ],
    },
    {
      key: 'grok',
      name: 'Grok',
      icon: ['svg', 'grok'],
      defaultModels: [
        'grok-3',
        'grok-3-fast',
        'grok-3-mini',
        'grok-3-mini-fast',
      ],
    },
    {
      key: 'qwen',
      name: 'Qwen',
      icon: ['svg', 'qwen'],
      defaultModels: [
        'qwen-max',
        'qwen-plus',
        'qwen-turbo',
        'qwq-plus',
        'qwq-32b',
        'qwen-omni-turbo',
      ],
    },
    {
      key: 'mistral',
      name: 'Mistral',
      icon: ['svg', 'mistral'],
      defaultModels: [
        'codestral-latest',
        'mistral-large-latest',
        'pixtral-large-latest',
        'mistral-saba-latest',
        'ministral-3b-latest',
        'ministral-8b-latest',
      ],
    },
    {
      key: 'deepseek',
      name: 'DeepSeek',
      icon: ['svg', 'deepseek'],
      defaultModels: ['deepseek-chat', 'deepseek-reasoner'],
    },
    {
      key: 'openai-compatible',
      name: 'OpenAI Compatible',
      icon: ['svg', 'openai'],
    },
  ];
};

export const getLLMProviderItem = (
  providerKey: LLMProviderKey
): LLMProviderItem | undefined => {
  const providerItems = getLLMProviderItems();
  return providerItems.find(item => item.key === providerKey);
};

export const getLLMProviderName = (
  providerKey: LLMProviderKey
): string | undefined => {
  return getLLMProviderItem(providerKey)?.name;
};

export const getLLMProviderIcon = (
  providerKey: LLMProviderKey
): Icon | undefined => {
  return getLLMProviderItem(providerKey)?.icon;
};
