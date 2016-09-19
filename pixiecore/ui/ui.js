'use strict';

class MachineList extends HTMLElement {
    attachedCallback() {
        console.log('Augmented!');
    }
}

class Machine extends HTMLElement {
    attachedCallback() {
        console.log('Augmented!');
    }
}

document.registerElement('px-machine-list', MachineList);
document.registerElement('px-machine', Machine);
