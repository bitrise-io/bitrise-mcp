import json
import re
from typing import Dict, List, Optional, Set


async def process_build_log(
    log_response: str,
    get_log_of_failed_step_only: bool = False,
    enable_log_filtering: bool = False,
    log_filter_patterns: Optional[Dict[str, List[str]]] = None,
    log_filter_context_lines: int = 5,
    failed_step_index: int = 1,
    show_summary_if_no_failures: bool = False,
    app_slug: str = None,
    build_slug: str = None
) -> str:
    """
    Process and compact build log response from Bitrise API.
    Uses expiring_raw_log_url for complete log content if available.
    
    Args:
        log_response: Raw JSON response from Bitrise build log API
        get_log_of_failed_step_only: Extract only failed step logs
        enable_log_filtering: Enable log filtering based on patterns
        log_filter_patterns: Dictionary mapping step types to keyword lists
        log_filter_context_lines: Number of context lines to include around matches
    
    Returns:
        Processed JSON response with compacted logs
    """
    try:
        log_data = json.loads(log_response)
    except json.JSONDecodeError:
        return log_response
    
    if not get_log_of_failed_step_only and not enable_log_filtering:
        return log_response
    
    # Use filter patterns directly (already a dict)
    filter_map = log_filter_patterns if enable_log_filtering and log_filter_patterns else {}
    
    # Get complete log content
    all_log_content = ""
    
    # Try to get full log from expiring_raw_log_url first
    if "expiring_raw_log_url" in log_data and log_data["expiring_raw_log_url"]:
        try:
            import httpx
            async with httpx.AsyncClient() as client:
                response = await client.get(log_data["expiring_raw_log_url"], timeout=30.0)
                if response.status_code == 200:
                    all_log_content = response.text
        except Exception:
            # Fall back to log_chunks if URL fetch fails
            pass
    
    # Fall back to log_chunks if no raw URL or fetch failed
    if not all_log_content and "log_chunks" in log_data:
        for chunk in log_data["log_chunks"]:
            all_log_content += chunk.get("chunk", "")
    
    if not all_log_content:
        return log_response
    
    # Process the complete log content
    failed_step_count = 0
    failed_step_names = []
    if get_log_of_failed_step_only and enable_log_filtering and filter_map:
        # Both enabled: first extract failed steps, then apply filtering
        failed_step_content, failed_step_count, failed_step_names = _extract_failed_step_logs(all_log_content, failed_step_index)
        # Only apply filtering if we have actual failed step content, not build summary
        if failed_step_count > 0:
            processed_content = _apply_content_filtering(failed_step_content, filter_map, log_filter_context_lines)
        else:
            # Return build summary without filtering
            processed_content = failed_step_content
    elif get_log_of_failed_step_only:
        processed_content, failed_step_count, failed_step_names = _extract_failed_step_logs(all_log_content, failed_step_index)
    elif enable_log_filtering and filter_map:
        processed_content = _apply_content_filtering(all_log_content, filter_map, log_filter_context_lines)
    else:
        processed_content = all_log_content
    
    # Create new response with note at the top
    processed_response = {}
    
    # Add note first
    if get_log_of_failed_step_only and failed_step_count > 0:
        if failed_step_count > 1:
            step_list = ", ".join([name for name, _ in failed_step_names])
            if failed_step_index <= failed_step_count:
                processed_response["note"] = f"Showing failed step #{failed_step_index} of {failed_step_count}. Total failed steps: {step_list}. Use compact_logs={failed_step_index + 1} for next failed step."
                # Add next command if we have the slugs and there are more steps
                if app_slug and build_slug and failed_step_index < failed_step_count:
                    processed_response["next_command"] = f"get_build_log('{app_slug}', '{build_slug}', {failed_step_index + 1})"
            else:
                processed_response["note"] = f"Requested failed step #{failed_step_index} but only {failed_step_count} failed steps exist. Showing build summary instead. All failed steps: {step_list}. Use compact_logs=1 to {failed_step_count} to see specific failed steps."
        else:
            processed_response["note"] = f"Showing the only failed step (step #{failed_step_index})"
    elif get_log_of_failed_step_only and failed_step_count == 0:
        processed_response["note"] = "No failed steps found. Showing build summary instead."
    elif enable_log_filtering:
        processed_response["note"] = "Logs filtered using keyword patterns"
    else:
        processed_response["note"] = "Log processing completed"
    
    # Add all original fields after note
    for key, value in log_data.items():
        processed_response[key] = value
    
    # Replace log_chunks with processed content
    processed_response["log_chunks"] = [{"chunk": processed_content, "position": 0}]
    
    # Remove the raw URL since we've processed the content
    if "expiring_raw_log_url" in processed_response:
        processed_response["expiring_raw_log_url"] = None
    
    return json.dumps(processed_response)




def _detect_workflow_structure(lines: List[str]) -> Dict:
    """Detect workflow structure: chained, nested, or both."""
    structure = {
        'is_chained': False,
        'chained_name': None,
        'chained_workflows': [],
        'has_nested': False,
        'chained_instance': None  # Pre-calculated chained workflow instance
    }
    
    # 1. Detect if it is chained (should show in the beginning)
    for line in lines:
        if "Running workflows:" in line and "→" in line:
            structure['is_chained'] = True
            clean_line = re.sub(r'\x1b\[[0-9;]*m', '', line)
            workflow_match = re.search(r'Running workflows:\s*(.+)', clean_line)
            if workflow_match:
                workflow_chain = workflow_match.group(1).strip()
                structure['chained_name'] = workflow_chain
                structure['chained_workflows'] = [w.strip() for w in workflow_chain.split('→')]
                # Pre-calculate the chained workflow instance identifier
                structure['chained_instance'] = f"{workflow_chain} [1]"
            break
    
    # 2. Check if multiple summaries exist (indicates nested workflows)
    summary_count = sum(1 for line in lines if "bitrise summary:" in line)
    structure['has_nested'] = summary_count > 1
    
    # 3. If both, then it is chained with nested
    return structure


def _extract_failed_step_logs(log_content: str, failed_step_index: int = 1) -> tuple[str, int, List[str]]:
    """Extract logs from failed steps with smart workflow handling."""
    lines = log_content.split('\n')
    
    # First, detect the overall workflow structure
    workflow_structure = _detect_workflow_structure(lines)
    
    # Then find failed steps from the summary
    failed_step_names, summary_content = _find_failed_steps_from_summary(lines, workflow_structure)
    
    if not failed_step_names:
        # No failed steps - return summary content if available, otherwise original content
        return summary_content if summary_content else log_content, 0, []
    
    # Parse steps with the detected workflow structure
    steps, workflow_hierarchy = _parse_bitrise_steps(lines, workflow_structure)
    failed_step_content = []
    
    # Extract step names with workflow context to avoid incorrect deduplication
    clean_step_names = []
    for failed_name, workflow_instance in failed_step_names:
        # Remove sequence numbers like [1] from step names
        clean_name = re.sub(r'\s*\[\d+\]$', '', failed_name)
        # Create unique identifier with workflow context to avoid incorrect deduplication
        # e.g., "Build Cache for Gradle (in unit-tests-no-shards)", "Build Cache for Gradle (in run_tests)"
        workflow_key = workflow_instance.split(' [')[0]  # Remove [1] from workflow
        step_with_context = f"{clean_name} (in {workflow_key})"
        clean_step_names.append(step_with_context)
    
    # Group failed steps by workflow to detect nested failures
    failed_by_workflow = {}
    for failed_name, failed_workflow_instance in failed_step_names:
        for step in steps:
            # Match step name with sequence number (all steps now have sequence numbers)
            name_matches = failed_name.lower() == step['title_with_count'].lower()
            
            # For chained workflows, check if step's workflow is part of the chain
            if " → " in failed_workflow_instance:
                # Extract workflow names from chained identifier
                chain_part = failed_workflow_instance.split(" [")[0]  # Remove [1] suffix
                chain_workflows = [w.strip() for w in chain_part.split(" → ")]
                workflow_matches = any(chain_wf in step['workflow'] for chain_wf in chain_workflows)
            else:
                # Regular workflow path matching
                workflow_matches = failed_workflow_instance in step['workflow_path']
            
            if name_matches and workflow_matches:
                workflow = step['workflow']
                if workflow not in failed_by_workflow:
                    failed_by_workflow[workflow] = []
                failed_by_workflow[workflow].append(step)
    
    # Apply smart filtering based on workflow structure
    included_steps = set()
    
    if workflow_structure and workflow_structure['is_chained'] and not workflow_structure['has_nested']:
        # For true chained workflows with single end summary, include all failed steps
        # No nested filtering needed since it's a sequential chain, not nested structure
        for workflow, workflow_failed_steps in failed_by_workflow.items():
            for step in workflow_failed_steps:
                included_steps.add(id(step))
    else:
        # For nested workflows or complex cases, apply smart filtering
        for workflow, workflow_failed_steps in failed_by_workflow.items():
            # Check if this workflow has child workflows with failures
            has_child_failures = False
            if workflow in workflow_hierarchy:
                for child_workflow in workflow_hierarchy[workflow]:
                    if child_workflow in failed_by_workflow:
                        has_child_failures = True
                        break
            
            if has_child_failures:
                # This workflow has nested failures - skip its own failed steps
                # The nested failed steps will be included instead
                continue
            else:
                # This workflow has direct failures - include them
                for step in workflow_failed_steps:
                    included_steps.add(id(step))
    
    # Extract content from included steps
    total_failed_steps = len(included_steps)
    steps_added = 0
    
    # If requested index exceeds available failed steps, return summary instead
    if failed_step_index > total_failed_steps:
        return summary_content if summary_content else log_content, len(failed_step_names), [(clean_name, '') for clean_name in clean_step_names]
    
    for step in steps:
        if id(step) in included_steps:
            steps_added += 1
            
            # Only include the step if it matches our desired index (1-based)
            if steps_added == failed_step_index:
                failed_step_content.extend(step['content'])
                failed_step_content.append("")  # Add separator
                break
    
    return ('\n'.join(failed_step_content) if failed_step_content else log_content, len(failed_step_names), [(clean_name, '') for clean_name in clean_step_names])


def _find_failed_steps_from_summary(lines: List[str], workflow_structure: Dict = None) -> tuple[List[tuple[str, str]], str]:
    """Find failed step names and their workflow instances from all Bitrise workflow summaries in the log.
    
    Returns:
        Tuple of (failed_steps, summary_content):
            - failed_steps: List of tuples (step_name, workflow_path_with_instance) 
            - summary_content: String containing all summary sections
    """
    failed_steps = []
    workflow_execution_count = {}  # Track workflow executions to match step parsing
    
    # Use workflow structure info passed from caller
    is_chained_workflow = workflow_structure and workflow_structure['is_chained']
    chained_workflow_names = workflow_structure['chained_workflows'] if workflow_structure else []
    chained_workflow_name = workflow_structure['chained_name'] if workflow_structure else None
    
    # Find all summary sections (there can be multiple with nested workflows)
    summary_sections = []
    
    for i, line in enumerate(lines):
        if "bitrise summary:" in line:
            # Extract workflow name from summary
            workflow_match = re.search(r'bitrise summary:\s+(.+?)\s*\|', line)
            if workflow_match:
                workflow_name = workflow_match.group(1).strip()
                
                
                # Find the end of this summary section
                summary_end = -1
                for j in range(i + 1, len(lines)):
                    if "Total runtime:" in lines[j]:
                        # Found the total runtime line, summary ends at the next border line
                        for k in range(j + 1, len(lines)):
                            if lines[k].startswith("+"):
                                summary_end = k
                                break
                        break
                    elif "bitrise summary:" in lines[j]:
                        # Another summary starts, this one ends
                        summary_end = j - 1
                        break
                
                if summary_end == -1:
                    summary_end = len(lines) - 1
                
                # Store temporarily - we'll assign workflow instance after collecting all summaries
                summary_sections.append((i, summary_end, workflow_name))
    
    # Assign workflow instances based on user's logic:
    # - if final summary and chained workflow exists: assign to chained workflow
    # - if final summary and no chained workflow: assign to current workflow  
    # - if not final summary: assign to current workflow
    summary_sections_with_instances = []
    for idx, (summary_start, summary_end, workflow_name) in enumerate(summary_sections):
        is_final_summary = (idx == len(summary_sections) - 1)
        
        if is_final_summary and is_chained_workflow:
            # Final summary in chained workflow - use pre-calculated instance
            workflow_instance = workflow_structure['chained_instance']
        else:
            # Not final summary, or final summary with no chained workflow - assign to current workflow
            if workflow_name not in workflow_execution_count:
                workflow_execution_count[workflow_name] = 0
            workflow_execution_count[workflow_name] += 1
            workflow_instance = f"{workflow_name} [{workflow_execution_count[workflow_name]}]"
        
        summary_sections_with_instances.append((summary_start, summary_end, workflow_instance))
    
    # Parse each summary section
    for summary_start, summary_end, workflow_instance in summary_sections_with_instances:
        summary_step_counts = {}  # Track step name occurrences within this summary
        
        for i in range(summary_start, summary_end + 1):
            line = lines[i]
            
            # Skip header and separator lines
            if line.startswith("+") or "title" in line or "bitrise summary" in line:
                continue
                
            # Parse table rows: | status | step name | time |
            if line.startswith("|") and line.count("|") >= 3:
                parts = [part.strip() for part in line.split("|")]
                if len(parts) >= 4:  # ['', status, step_name, time, '']
                    status_part = parts[1]
                    step_name_part = parts[2] 
                    
                    # Clean up the step name first (remove ANSI codes and status suffixes)
                    clean_step_name = re.sub(r'\x1b\[[\d;]*m', '', step_name_part)  # Remove ANSI escape codes
                    clean_step_name = re.sub(r'\s*\((Failed|Skipped)\).*', '', clean_step_name)  # Remove status suffixes
                    clean_step_name = clean_step_name.strip()
                    
                    if clean_step_name:
                        # Track the occurrence of this step name in the summary
                        if clean_step_name not in summary_step_counts:
                            summary_step_counts[clean_step_name] = 0
                        summary_step_counts[clean_step_name] += 1
                        
                        # Create step identifier with sequence number (always add sequence number)
                        step_sequence = summary_step_counts[clean_step_name]
                        step_with_sequence = f"{clean_step_name} [{step_sequence}]"
                        
                        # Check if this specific occurrence failed
                        if ('x' in status_part and '31;1m' in status_part) or '(Failed)' in step_name_part:
                            step_with_workflow = (step_with_sequence, workflow_instance)
                            if step_with_workflow not in failed_steps:
                                failed_steps.append(step_with_workflow)
    
    # Get the final summary content (last summary section)
    final_summary_content = ""
    if summary_sections:
        final_start, final_end, final_workflow = summary_sections[-1]  # Get the last summary
        final_summary_content = '\n'.join(lines[final_start:final_end + 1])
    
    return failed_steps, final_summary_content


def _parse_bitrise_steps(lines: List[str], workflow_structure: Dict = None) -> tuple[List[Dict], Dict]:
    """Parse Bitrise log into individual steps with workflow context and hierarchy.
    
    Returns:
        tuple: (steps, workflow_hierarchy)
            - steps: List of step dictionaries with workflow context
            - workflow_hierarchy: Dict mapping parent workflows to child workflows
    """
    steps = []
    current_step = None
    workflow_stack = ["main"]  # Stack to track nested workflows
    workflow_hierarchy = {}  # Track parent-child relationships
    workflow_execution_count = {}  # Track how many times each workflow has been called
    workflow_step_counts = {}  # Track step name repetitions within each workflow
    
    # Use detected workflow structure
    is_chained_workflow = workflow_structure['is_chained']
    chained_workflow_name = workflow_structure['chained_name']
    
    # Use pre-calculated chained workflow instance
    chained_workflow_instance = workflow_structure['chained_instance'] if is_chained_workflow else None
    
    for i, line in enumerate(lines):
        # Check for workflow switching
        workflow_match = re.search(r'Switching to workflow:\s+(.+)', line)
        if workflow_match:
            workflow_name = workflow_match.group(1).strip()
            parent_workflow = workflow_stack[-1]
            
            # Track workflow execution count
            if workflow_name not in workflow_execution_count:
                workflow_execution_count[workflow_name] = 0
            workflow_execution_count[workflow_name] += 1
            
            # Create unique workflow identifier with execution count
            workflow_instance = f"{workflow_name} [{workflow_execution_count[workflow_name]}]"
            
            # Record hierarchy relationship
            if parent_workflow not in workflow_hierarchy:
                workflow_hierarchy[parent_workflow] = []
            if workflow_instance not in workflow_hierarchy[parent_workflow]:
                workflow_hierarchy[parent_workflow].append(workflow_instance)
            
            workflow_stack.append(workflow_instance)
            continue
            
        # Check for workflow summary (end of workflow)
        summary_match = re.search(r'bitrise summary:\s+(.+)', line)
        if summary_match:
            # Pop from workflow stack when we see a summary
            if len(workflow_stack) > 1:
                workflow_stack.pop()
            continue
        
        # Look for step header pattern: | (N) Step Name |
        step_match = re.match(r'^\|\s*\((\d+)\)\s+(.*?)\s*\|$', line)
        if step_match:
            # Save previous step
            if current_step:
                steps.append(current_step)
            
            # Start new step
            step_number = step_match.group(1)
            step_title = step_match.group(2)
            current_workflow = workflow_stack[-1]
            
            # For chained workflows, determine if this is a nested execution
            if is_chained_workflow and chained_workflow_instance:
                # Check if current workflow matches the first workflow in the chain (nested execution)
                chained_names = chained_workflow_name.split(' → ') if chained_workflow_name else []
                current_workflow_base = current_workflow.split(' [')[0] if ' [' in current_workflow else current_workflow
                
                if chained_names and current_workflow_base == chained_names[0]:
                    # This is a nested execution of the first workflow in the chain
                    effective_workflow = current_workflow
                else:
                    # This is a main chained workflow step
                    effective_workflow = chained_workflow_instance
            else:
                effective_workflow = current_workflow
            
            # Track step execution count within this effective workflow
            if effective_workflow not in workflow_step_counts:
                workflow_step_counts[effective_workflow] = {}
            
            if step_title not in workflow_step_counts[effective_workflow]:
                workflow_step_counts[effective_workflow][step_title] = 0
            workflow_step_counts[effective_workflow][step_title] += 1
            
            # Create unique step title with execution count within workflow (always add sequence number)
            step_count = workflow_step_counts[effective_workflow][step_title]
            step_title_with_count = f"{step_title} [{step_count}]"
            
            current_step = {
                'number': int(step_number),
                'title': step_title,
                'title_with_count': step_title_with_count,
                'workflow': effective_workflow,
                'workflow_path': " -> ".join(workflow_stack),  # Full hierarchy path
                'workflow_depth': len(workflow_stack) - 1,      # Nesting depth (0 = main)
                'parent_workflow': workflow_stack[-2] if len(workflow_stack) > 1 else None,
                'content': [line],
                'start_line': i,
                'workflow_at_start': current_workflow  # Track workflow when step started
            }
            continue
        
        # Add content to current step
        if current_step:
            current_step['content'].append(line)
            
            # Workflow-aware step boundary detection
            next_line_idx = i + 1
            if next_line_idx < len(lines):
                next_line = lines[next_line_idx]
                next_step_match = re.match(r'^\|\s*\(\d+\)\s+.*\s*\|$', next_line)
                if next_step_match:
                    # Next step detected - but only end current step if we're in the same workflow
                    # as when the current step started
                    if current_workflow == current_step['workflow_at_start']:
                        steps.append(current_step)
                        current_step = None
    
    # Save last step
    if current_step:
        steps.append(current_step)
    
    return steps, workflow_hierarchy


def _apply_content_filtering(log_content: str, filter_map: Dict[str, List[str]], context_lines: int = 5) -> str:
    """Apply keyword filtering to log content based on step titles."""
    lines = log_content.split('\n')
    
    # Detect workflow structure and parse steps
    workflow_structure = _detect_workflow_structure(lines)
    steps, workflow_hierarchy = _parse_bitrise_steps(lines, workflow_structure)
    filtered_content = []
    
    for step in steps:
        step_title = step['title'].lower()
        step_content_lines = step['content']
        
        # Find matching keywords for this step based on its title
        matching_keywords = []
        for step_type, keywords in filter_map.items():
            if step_type.lower() in step_title:
                matching_keywords = keywords
                break
        
        if matching_keywords:
            # Apply filtering to this step's content
            filtered_step_content = _filter_step_content(step_content_lines, matching_keywords, context_lines)
            filtered_content.extend(filtered_step_content)
        else:
            # No matching keywords, include the whole step
            filtered_content.extend(step_content_lines)
        
        filtered_content.append("")  # Add separator between steps
    
    return '\n'.join(filtered_content)


def _filter_step_content(step_lines: List[str], keywords: List[str], context_lines: int = 5) -> List[str]:
    """Filter a single step's content based on keywords."""
    if not keywords:
        return step_lines
    
    filtered_lines = []
    included_indices = set()
    
    # Always include step header lines (first ~10 lines typically contain step boundary info)
    step_header_lines = min(10, len(step_lines))
    for i in range(step_header_lines):
        line = step_lines[i]
        # Include step boundary markers, step info, and metadata
        if (line.startswith('|') or line.startswith('+') or 
            'id:' in line or 'version:' in line or 'collection:' in line or 'toolkit:' in line):
            included_indices.add(i)
    
    # Find lines matching keywords
    for i, line in enumerate(step_lines):
        for keyword in keywords:
            if keyword.lower() in line.lower():
                # Include this line and context lines above and below
                for j in range(max(0, i-context_lines), min(len(step_lines), i+context_lines+1)):
                    included_indices.add(j)
                break
    
    # Build filtered step content
    if included_indices:
        sorted_indices = sorted(included_indices)
        for i in sorted_indices:
            filtered_lines.append(step_lines[i])
        return filtered_lines
    
    # If no matches found, include step header only (first few lines)
    return step_lines[:5] if len(step_lines) > 5 else step_lines




