from flask import Blueprint, request, jsonify
import spacy

chatbot_bp = Blueprint('chatbot', __name__)
nlp = spacy.load('en_core_web_sm')  # Requires `python -m spacy download en_core_web_sm`

@chatbot_bp.route('/api/chatbot', methods=['POST'])
def chatbot():
    try:
        message = request.json['message']
        doc = nlp(message)
        # Simple intent detection (expand as needed)
        response = "I’m here to help! What do you need?" if "help" in message.lower() else "Can you clarify your request?"
        return jsonify({'response': response}), 200
    except Exception as e:
        return jsonify({'error': f"Chatbot error: {str(e)}"}), 400
