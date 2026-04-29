const API_BASE_URL = 'http://localhost:8080/feedback';

// Dark Mode Toggle
const darkModeToggle = document.getElementById('darkModeToggle');
darkModeToggle.addEventListener('click', () => {
    document.body.classList.toggle('dark-mode');
    const isDarkMode = document.body.classList.contains('dark-mode');
    darkModeToggle.textContent = isDarkMode ? '☀️ Light Mode' : '🌙 Dark Mode';
    localStorage.setItem('darkMode', isDarkMode);
});

// Load dark mode preference
if (localStorage.getItem('darkMode') === 'true') {
    document.body.classList.add('dark-mode');
    darkModeToggle.textContent = '☀️ Light Mode';
}

// Submit Feedback
document.getElementById('feedbackForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const feedback = {
        name: document.getElementById('name').value,
        email: document.getElementById('email').value,
        rating: parseInt(document.getElementById('rating').value),
        comments: document.getElementById('comments').value
    };
    
    try {
        const response = await fetch(API_BASE_URL, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(feedback)
        });
        
        if (response.ok) {
            alert('Feedback submitted successfully!');
            document.getElementById('feedbackForm').reset();
            loadAllFeedback();
        } else {
            const error = await response.json();
            alert('Error: ' + (error.error || 'Failed to submit feedback'));
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Failed to connect to server. Make sure the API is running.');
    }
});

// Load All Feedback
async function loadAllFeedback() {
    try {
        const response = await fetch(API_BASE_URL);
        if (response.ok) {
            const feedbacks = await response.json();
            displayFeedback(feedbacks);
            setActiveButton('All');
        } else {
            showError('Failed to load feedback');
        }
    } catch (error) {
        console.error('Error:', error);
        showError('Failed to connect to server');
    }
}

// Load Only Names
async function loadOnlyNames() {
    try {
        const response = await fetch(`${API_BASE_URL}/names`);
        if (response.ok) {
            const names = await response.json();
            displaySimpleList(names, 'Names');
            setActiveButton('Only Names');
        } else {
            showError('Failed to load names');
        }
    } catch (error) {
        console.error('Error:', error);
        showError('Failed to connect to server');
    }
}

// Load Only Emails
async function loadOnlyEmails() {
    try {
        const response = await fetch(`${API_BASE_URL}/emails`);
        if (response.ok) {
            const emails = await response.json();
            displaySimpleList(emails, 'Emails');
            setActiveButton('Only Emails');
        } else {
            showError('Failed to load emails');
        }
    } catch (error) {
        console.error('Error:', error);
        showError('Failed to connect to server');
    }
}

// Load Only Comments
async function loadOnlyComments() {
    try {
        const response = await fetch(`${API_BASE_URL}/comments`);
        if (response.ok) {
            const comments = await response.json();
            displaySimpleList(comments, 'Messages');
            setActiveButton('Only Messages');
        } else {
            showError('Failed to load messages');
        }
    } catch (error) {
        console.error('Error:', error);
        showError('Failed to connect to server');
    }
}

// Display Full Feedback
function displayFeedback(feedbacks) {
    const container = document.getElementById('feedbackList');
    
    if (!feedbacks || feedbacks.length === 0) {
        container.innerHTML = '<div class="no-data">No feedback found. Be the first to submit!</div>';
        return;
    }
    
    container.innerHTML = feedbacks.map(feedback => `
        <div class="feedback-item" data-id="${feedback.id}">
            <h3>${escapeHtml(feedback.name)}</h3>
            <div class="rating">${getStarRating(feedback.rating)}</div>
            <p><strong>Message:</strong> ${escapeHtml(feedback.comments || 'No message')}</p>
            <div class="email">📧 ${escapeHtml(feedback.email)}</div>
            <div class="date">📅 ${new Date(feedback.created_at).toLocaleString()}</div>
            <div class="feedback-actions">
                <button onclick="openUpdateModal(${feedback.id})" class="update-btn">Update</button>
                <button onclick="deleteFeedback(${feedback.id})" class="delete-btn">Delete</button>
            </div>
        </div>
    `).join('');
}

// Display Simple List (Names, Emails, Comments)
function displaySimpleList(items, title) {
    const container = document.getElementById('feedbackList');
    
    if (!items || items.length === 0) {
        container.innerHTML = `<div class="no-data">No ${title.toLowerCase()} found</div>`;
        return;
    }
    
    container.innerHTML = `
        <div class="simple-list">
            <h3>${title} (${items.length})</h3>
            ${items.map(item => `<div class="simple-item">${escapeHtml(item)}</div>`).join('')}
        </div>
    `;
}

// Update Feedback
async function updateFeedback(id) {
    const name = document.getElementById('updateName').value;
    const email = document.getElementById('updateEmail').value;
    const rating = parseInt(document.getElementById('updateRating').value);
    const comments = document.getElementById('updateComments').value;
    
    try {
        const response = await fetch(`${API_BASE_URL}/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ name, email, rating, comments })
        });
        
        if (response.ok) {
            alert('Feedback updated successfully!');
            closeModal();
            loadAllFeedback();
        } else {
            const error = await response.json();
            alert('Error: ' + (error.error || 'Failed to update feedback'));
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Failed to connect to server');
    }
}

// Delete Feedback
async function deleteFeedback(id) {
    if (confirm('Are you sure you want to delete this feedback?')) {
        try {
            const response = await fetch(`${API_BASE_URL}/${id}`, {
                method: 'DELETE'
            });
            
            if (response.ok) {
                alert('Feedback deleted successfully!');
                loadAllFeedback();
            } else {
                const error = await response.json();
                alert('Error: ' + (error.error || 'Failed to delete feedback'));
            }
        } catch (error) {
            console.error('Error:', error);
            alert('Failed to connect to server');
        }
    }
}

// Open Update Modal
async function openUpdateModal(id) {
    try {
        const response = await fetch(`${API_BASE_URL}/${id}`);
        if (response.ok) {
            const feedback = await response.json();
            document.getElementById('updateId').value = feedback.id;
            document.getElementById('updateName').value = feedback.name;
            document.getElementById('updateEmail').value = feedback.email;
            document.getElementById('updateRating').value = feedback.rating;
            document.getElementById('updateComments').value = feedback.comments || '';
            document.getElementById('updateModal').style.display = 'block';
        } else {
            alert('Failed to load feedback details');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Failed to connect to server');
    }
}

// Helper Functions
function getStarRating(rating) {
    const fullStar = '⭐';
    const emptyStar = '☆';
    return fullStar.repeat(rating) + emptyStar.repeat(5 - rating);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showError(message) {
    const container = document.getElementById('feedbackList');
    container.innerHTML = `<div class="error">❌ ${message}</div>`;
}

function setActiveButton(activeText) {
    const buttons = document.querySelectorAll('.filter-btn');
    buttons.forEach(btn => {
        if (btn.textContent === activeText) {
            btn.classList.add('active');
        } else {
            btn.classList.remove('active');
        }
    });
}

function closeModal() {
    document.getElementById('updateModal').style.display = 'none';
}

// Modal HTML (add to body)
const modalHTML = `
    <div id="updateModal" class="modal">
        <div class="modal-content">
            <span class="close" onclick="closeModal()">&times;</span>
            <h2>Update Feedback</h2>
            <input type="hidden" id="updateId">
            <div class="form-group">
                <label>Name:</label>
                <input type="text" id="updateName" required>
            </div>
            <div class="form-group">
                <label>Email:</label>
                <input type="email" id="updateEmail" required>
            </div>
            <div class="form-group">
                <label>Rating:</label>
                <select id="updateRating" required>
                    <option value="5">⭐⭐⭐⭐⭐ - Excellent</option>
                    <option value="4">⭐⭐⭐⭐ - Good</option>
                    <option value="3">⭐⭐⭐ - Average</option>
                    <option value="2">⭐⭐ - Poor</option>
                    <option value="1">⭐ - Very Poor</option>
                </select>
            </div>
            <div class="form-group">
                <label>Message:</label>
                <textarea id="updateComments" rows="3"></textarea>
            </div>
            <button onclick="updateFeedback(document.getElementById('updateId').value)" class="submit-btn">Update</button>
        </div>
    </div>
`;

document.body.insertAdjacentHTML('beforeend', modalHTML);

// Load feedback on page load
loadAllFeedback();