export function applyRules(files, rules) {
    // Deep copy files to avoid mutating original
    const newFiles = JSON.parse(JSON.stringify(files));

    // Helper to process a single node
    function processNode(node) {
        let newName = node.name;

        for (const rule of rules) {
            if (!rule.enabled) continue;

            switch (rule.type) {
                case 'replace':
                    if (rule.params.find) {
                        newName = newName.split(rule.params.find).join(rule.params.replace || '');
                    }
                    break;
                case 'regex':
                    try {
                        if (rule.params.pattern) {
                            const re = new RegExp(rule.params.pattern, rule.params.flags || '');
                            newName = newName.replace(re, rule.params.replace || '');
                        }
                    } catch (e) {
                        console.error('Invalid regex:', e);
                    }
                    break;
                case 'prefix':
                    if (rule.params.text) {
                        newName = rule.params.text + newName;
                    }
                    break;
                case 'suffix':
                    if (rule.params.text) {
                        const parts = newName.split('.');
                        if (parts.length > 1) {
                            const ext = parts.pop();
                            newName = parts.join('.') + rule.params.text + '.' + ext;
                        } else {
                            newName = newName + rule.params.text;
                        }
                    }
                    break;
                case 'extension':
                    if (rule.params.mode === 'lowercase') {
                        const parts = newName.split('.');
                        if (parts.length > 1) {
                            const ext = parts.pop();
                            newName = parts.join('.') + '.' + ext.toLowerCase();
                        }
                    } else if (rule.params.mode === 'uppercase') {
                        const parts = newName.split('.');
                        if (parts.length > 1) {
                            const ext = parts.pop();
                            newName = parts.join('.') + '.' + ext.toUpperCase();
                        }
                    }
                    break;
                case 'date_prefix':
                    if (node.mod_time) {
                        const date = new Date(node.mod_time);
                        const year = date.getFullYear();
                        const month = String(date.getMonth() + 1).padStart(2, '0');
                        const day = String(date.getDate()).padStart(2, '0');
                        const prefix = `${year}${month}${day}_`;

                        if (!newName.startsWith(prefix)) {
                            newName = prefix + newName;
                        }
                    }
                    break;
            }
        }

        node.name = newName;

        if (node.children) {
            node.children.forEach(processNode);
        }
    }

    newFiles.forEach(processNode);
    return newFiles;
}
