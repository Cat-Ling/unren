#!/usr/bin/env python3
"""
Extract Python tools from UnRen batch file and prepare them for embedding in Go.
This script decodes the Base64-encoded rpatool and unrpyc from the batch file.
"""

import base64
import re
import os
import sys

def extract_vars_from_batch(batch_path: str) -> dict:
    """Extract set VAR=VALUE pairs from batch file."""
    variables = {}
    with open(batch_path, 'r', encoding='utf-8', errors='ignore') as f:
        for line in f:
            # Match: set varname=value or REM set varname=value
            match = re.match(r'^(?:REM\s+)?set\s+(\w+)=(.*)$', line.strip())
            if match:
                var_name = match.group(1)
                var_value = match.group(2).rstrip('\r\n')
                variables[var_name] = var_value
    return variables

def decode_and_save(variables: dict, var_prefixes: list, output_path: str):
    """Concatenate Base64 chunks and decode to file."""
    # Collect all matching vars in order
    chunks = []
    for prefix in var_prefixes:
        # Find all vars with this prefix, sorted by number
        matching = [(k, v) for k, v in variables.items() if k.startswith(prefix)]
        matching.sort(key=lambda x: int(re.search(r'\d+', x[0]).group()) if re.search(r'\d+', x[0]) else 0)
        for k, v in matching:
            chunks.append(v)
    
    if not chunks:
        print(f"  Warning: No variables found for prefixes {var_prefixes}")
        return False
    
    # Join and decode
    b64_data = ''.join(chunks)
    try:
        decoded = base64.b64decode(b64_data)
        os.makedirs(os.path.dirname(output_path), exist_ok=True)
        with open(output_path, 'wb') as f:
            f.write(decoded)
        print(f"  Saved: {output_path} ({len(decoded)} bytes)")
        return True
    except Exception as e:
        print(f"  Error decoding: {e}")
        return False

def main():
    if len(sys.argv) < 2:
        print("Usage: extract_python_tools.py <path_to_unren.bat>")
        sys.exit(1)
    
    batch_path = sys.argv[1]
    output_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    python_dir = os.path.join(output_dir, 'files', 'python')
    
    print(f"Extracting Python tools from: {batch_path}")
    print(f"Output directory: {python_dir}")
    
    # Parse batch file
    variables = extract_vars_from_batch(batch_path)
    print(f"Found {len(variables)} variables in batch file")
    
    # Extract rpatool Python 2 version (rpatool01-06)
    print("\n=== rpatool (Python 2) ===")
    decode_and_save(variables, ['rpatool0'], os.path.join(python_dir, 'rpatool_py2.py'))
    
    # Extract rpatool Python 3 version (rpatool07-12)
    # Actually, looking at the batch file, 07-12 is the Python 3 version
    # But they're named the same way, let me check the actual content
    
    # The batch file shows:
    # - rpatool01-06: Python 2/3 compatible (starts with "#!/usr/bin/env python")
    # - rpatool07-12: Python 3 specific (starts with "#!/usr/bin/env python3")
    print("\n=== rpatool (Python 3) ===")
    # rpatool07-12 are the Python 3 specific version
    py3_chunks = []
    for i in range(7, 13):
        key = f'rpatool{i:02d}'
        if key in variables:
            py3_chunks.append((key, variables[key]))
    
    if py3_chunks:
        b64_data = ''.join([v for k, v in py3_chunks])
        try:
            decoded = base64.b64decode(b64_data)
            output_path = os.path.join(python_dir, 'rpatool.py')
            os.makedirs(os.path.dirname(output_path), exist_ok=True)
            with open(output_path, 'wb') as f:
                f.write(decoded)
            print(f"  Saved: {output_path} ({len(decoded)} bytes)")
        except Exception as e:
            print(f"  Error: {e}")
    
    # Extract rpa.py fallback (rpatool20)
    print("\n=== rpa.py (fallback extractor) ===")
    if 'rpatool20' in variables:
        try:
            decoded = base64.b64decode(variables['rpatool20'])
            output_path = os.path.join(python_dir, 'rpa.py')
            with open(output_path, 'wb') as f:
                f.write(decoded)
            print(f"  Saved: {output_path} ({len(decoded)} bytes)")
        except Exception as e:
            print(f"  Error: {e}")
    
    # Extract unrpyc CAB files (decompcab40-53 for Python 3)
    print("\n=== unrpyc (Python 3 CAB) ===")
    cab_chunks = []
    for i in range(40, 54):
        key = f'decompcab{i}'
        if key in variables:
            cab_chunks.append((key, variables[key]))
    
    if cab_chunks:
        b64_data = ''.join([v for k, v in cab_chunks])
        try:
            decoded = base64.b64decode(b64_data)
            output_path = os.path.join(python_dir, 'unrpyc_py3.cab')
            with open(output_path, 'wb') as f:
                f.write(decoded)
            print(f"  Saved: {output_path} ({len(decoded)} bytes)")
        except Exception as e:
            print(f"  Error: {e}")
    
    # Extract unrpyc CAB files (decompcab20-33 for Python 2)
    print("\n=== unrpyc (Python 2 CAB) ===")
    cab_chunks = []
    for i in range(20, 34):
        key = f'decompcab{i}'
        if key in variables:
            cab_chunks.append((key, variables[key]))
    
    if cab_chunks:
        b64_data = ''.join([v for k, v in cab_chunks])
        try:
            decoded = base64.b64decode(b64_data)
            output_path = os.path.join(python_dir, 'unrpyc_py2.cab')
            with open(output_path, 'wb') as f:
                f.write(decoded)
            print(f"  Saved: {output_path} ({len(decoded)} bytes)")
        except Exception as e:
            print(f"  Error: {e}")
    
    print("\n=== Extraction complete ===")

if __name__ == '__main__':
    main()
