# Project Refinement Task

Please execute the following tasks to improve the forum application.

<bug_report>
    <title>Post Count Discrepancy</title>
    <description>
        When logged in as **erti@erti.com** (pass: **ertierti**), the user-card counter displays **2 posts**, but clicking the 'my posts' button reveals **only one post**.
    </description>
    <action_items>
        <item>Verify the issue using `curl` to check the API response or Playwright for visual verification.</item>
        <item>Debug the SQL queries or service logic responsible for counting posts vs retrieving the 'my posts' list.</item>
        <item>Fix the bug to ensure the count matches the displayed posts.</item>
    </action_items>
</bug_report>

<ui_enhancement>
    <target>Post Cards (Board Page)</target>
    <instruction>
        Make the **middle section** of the post cards on the board page **larger**. This likely involves adjusting the flex-grow or width properties in the CSS grid/flex layout.
    </instruction>
</ui_enhancement>

<ui_enhancement>
    <target>Reaction Buttons and Category Tags (Board Page)</target>
    <instruction>
        Make the **reaction buttons** and **category tags** **Closer to Each Other**. The space between the groups should be larger and the distance between the tags should be smaller.
    </instruction>
</ui_enhancement>

<layout_adjustment>
    <target>Sidebar / Main Layout</target>
    <instruction>
        **Swap the column positions**:
        1. Place the **User Card** column on the **LEFT** side of the screen.
        2. Place the other sidebar elements (Filters, Image, Categories widget, etc.) on the **RIGHT** side (or separate them if they were grouped).
    </instruction>
</layout_adjustment>

<bug_report>
    <title>User Card Visibility</title>
    <description>
        The **user-card** is currently missing when the user navigates to the **Create Post** or **Edit Post** pages.
    </description>
    <expected_behavior>
        The user-card must **ALWAYS** be visible when the user is logged in, regardless of the page (including Create/Edit views).
    </expected_behavior>
</bug_report>
