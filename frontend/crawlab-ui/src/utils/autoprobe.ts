export const getIconBySelectorType = (selectorType: SelectorType): Icon => {
  switch (selectorType) {
    case 'css':
      return ['fab', 'css'];
    case 'xpath':
      return ['fa', 'code'];
    case 'regex':
      return ['fa', 'search'];
  }
};

export const getIconByExtractType = (extractType?: ExtractType): Icon => {
  switch (extractType) {
    case 'text':
      return ['fa', 'file-alt'];
    case 'attribute':
      return ['fa', 'tag'];
    case 'html':
      return ['fa', 'file-code'];
    default:
      return ['fa', 'question'];
  }
};

export const getIconByItemType = (itemType?: AutoProbeItemType): Icon => {
  switch (itemType) {
    case 'page_pattern':
      return ['fa', 'network-wired'];
    case 'field':
      return ['fa', 'tag'];
    case 'list':
      return ['fa', 'list'];
    case 'pagination':
      return ['fa', 'ellipsis-h'];
    default:
      return ['fa', 'question'];
  }
};
