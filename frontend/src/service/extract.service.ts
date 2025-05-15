import { Injectable } from '@angular/core';
import { Extract } from '../../wailsjs/go/main/App';

@Injectable({
  providedIn: 'root'
})
export class ExtractService {
  async extraction(file: string) {
    await Extract(file);
  }
}
